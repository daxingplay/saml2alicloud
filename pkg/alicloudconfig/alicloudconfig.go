package alicloudconfig

import (
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"time"

	config "github.com/aliyun/aliyun-cli/config"
	homedir "github.com/mitchellh/go-homedir"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	ini "gopkg.in/ini.v1"
)

var (
	// ErrCredentialsHomeNotFound returned when a user home directory can't be located.
	ErrCredentialsHomeNotFound = errors.New("user home directory not found")

	// ErrCredentialsNotFound returned when the required aws credentials don't exist.
	ErrCredentialsNotFound = errors.New("alicloud credentials not found")

	logger = logrus.WithField("pkg", "alicloudconfig")
)

// AliCloudCredentials represents the set of attributes used to authenticate to AlibabaCloud with a short lived session
type AliCloudCredentials struct {
	AliCloudAccessKey     string    `ini:"aws_access_key_id"`
	AliCloudSecretKey     string    `ini:"aws_secret_access_key"`
	AliCloudSessionToken  string    `ini:"aws_session_token"`
	AliCloudSecurityToken string    `ini:"aws_security_token"`
	PrincipalARN          string    `ini:"x_principal_arn"`
	Expires               time.Time `ini:"x_security_token_expires"`
	Region                string    `ini:"region,omitempty"`
}

// CredentialsProvider loads aws credentials file
type CredentialsProvider struct {
	Filename string
	Profile  string
}

// NewSharedCredentials helper to create the credentials provider
func NewSharedCredentials(profile string) *CredentialsProvider {
	return &CredentialsProvider{
		Profile: profile,
	}
}

// CredsExists verify that the credentials exist
func (p *CredentialsProvider) CredsExists() (bool, error) {
	filename, err := p.resolveFilename()
	if err != nil {
		return false, err
	}

	err = p.ensureConfigExists()
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, errors.Wrapf(err, "unable to load file %s", filename)
	}

	return true, nil
}

// Save persist the credentials
func (p *CredentialsProvider) Save(awsCreds *AliCloudCredentials) error {
	configuration, err := config.LoadConfiguration(config.GetConfigPath()+"/config.json", os.Stdout)
	if err != nil {
		return err
	}
	profile := config.Profile{
		Name:            p.Profile,
		Mode:            config.StsToken,
		AccessKeyId:     awsCreds.AliCloudAccessKey,
		AccessKeySecret: awsCreds.AliCloudSecretKey,
		StsToken:        awsCreds.AliCloudSecurityToken,
		OutputFormat:    "json",
		Language:        "en",
	}
	configuration.PutProfile(profile)
	return config.SaveConfiguration(configuration)
}

// Load load the aws credentials file
func (p *CredentialsProvider) Load() (*config.Profile, error) {
	_, err := p.resolveFilename()
	if err != nil {
		return nil, err
	}

	profile, err := config.LoadCurrentProfile(os.Stdout)
	if err != nil {
		return nil, err
	}

	return &profile, nil
}

// Expired checks if the current credentials are expired
func (p *CredentialsProvider) Expired() bool {
	// _, err := p.Load()
	// if err != nil {
	// 	return true
	// }

	return true
}

// ensureConfigExists verify that the config file exists
func (p *CredentialsProvider) ensureConfigExists() error {
	filename, err := p.resolveFilename()
	if err != nil {
		return err
	}
	logger.WithField("filename", filename).Debug("ensureConfigExists")

	if _, err := os.Stat(filename); err != nil {
		if os.IsNotExist(err) {

			dir := filepath.Dir(filename)

			err = os.MkdirAll(dir, os.ModePerm)
			if err != nil {
				return err
			}

			logger.WithField("dir", dir).Debug("Dir created")

			// create an base config file
			err = ioutil.WriteFile(filename, []byte("["+p.Profile+"]"), 0600)
			if err != nil {
				return err
			}

			logger.WithField("filename", filename).Debug("File created")

		}
		return err
	}

	return nil
}

func (p *CredentialsProvider) resolveFilename() (string, error) {
	if p.Filename == "" {
		filename, err := locateConfigFile()
		if err != nil {
			return "", err
		}

		p.Filename = filename
	}

	return p.Filename, nil
}

func locateConfigFile() (string, error) {

	var name string
	var err error
	if runtime.GOOS == "windows" {
		name = path.Join(os.Getenv("USERPROFILE"), ".aliyun", "config.json")
	} else {
		name, err = homedir.Expand("~/.aliyun/config.json")
		if err != nil {
			return "", ErrCredentialsHomeNotFound
		}
	}
	logger.WithField("name", name).Debug("Expand")

	// is the filename a symlink?
	name, err = resolveSymlink(name)
	if err != nil {
		return "", errors.Wrap(err, "unable to resolve symlink")
	}

	logger.WithField("name", name).Debug("resolveSymlink")

	return name, nil
}

func resolveSymlink(filename string) (string, error) {
	sympath, err := filepath.EvalSymlinks(filename)

	// return the un modified filename
	if os.IsNotExist(err) {
		return filename, nil
	}
	if err != nil {
		return "", err
	}

	return sympath, nil
}

func createAndSaveProfile(filename, profile string, awsCreds *AliCloudCredentials) error {

	dirPath := filepath.Dir(filename)

	err := os.Mkdir(dirPath, 0700)
	if err != nil {
		return errors.Wrapf(err, "unable to create %s directory", dirPath)
	}

	_, err = os.Create(filename)
	if err != nil {
		return errors.Wrapf(err, "unable to create configuration")
	}

	return saveProfile(filename, profile, awsCreds)
}

func saveProfile(filename, profile string, awsCreds *AliCloudCredentials) error {
	config, err := ini.Load(filename)
	if err != nil {
		return err
	}
	iniProfile, err := config.NewSection(profile)
	if err != nil {
		return err
	}

	err = iniProfile.ReflectFrom(awsCreds)
	if err != nil {
		return err
	}

	return config.SaveTo(filename)
}
