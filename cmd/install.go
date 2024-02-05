package cmd

import (
	"fmt"
	"io"
	"os"
	"path"

	systemd "github.com/forge4flow/forge4flow-manager/pkg/systemd"
	"github.com/pkg/errors"

	"github.com/spf13/cobra"
)

var installCmd = &cobra.Command{
	Use:   "install",
	Short: "Install Forge4Flow-Manager",
	RunE:  runInstall,
}

const workingDirectoryPermission = 0644

const f4fManagerWd = "/var/lib/f4f-manager"

const faasdProviderWd = "/var/lib/faasd-provider"

func runInstall(_ *cobra.Command, _ []string) error {

	if err := ensureWorkingDir(path.Join(f4fManagerWd, "secrets")); err != nil {
		return err
	}

	if err := ensureWorkingDir(faasdProviderWd); err != nil {
		return err
	}

	if basicAuthErr := makeBasicAuthFiles(path.Join(f4fManagerWd, "secrets")); basicAuthErr != nil {
		return errors.Wrap(basicAuthErr, "cannot create basic-auth-* files")
	}

	if err := cp("docker-compose.yaml", f4fManagerWd); err != nil {
		return err
	}

	if err := cp("prometheus.yml", f4fManagerWd); err != nil {
		return err
	}

	if err := cp("resolv.conf", f4fManagerWd); err != nil {
		return err
	}

	err := binExists("/usr/local/bin/", "f4f-manager")
	if err != nil {
		return err
	}

	err = systemd.InstallUnit("faasd-provider", map[string]string{
		"Cwd":             faasdProviderWd,
		"SecretMountPath": path.Join(f4fManagerWd, "secrets")})

	if err != nil {
		return err
	}

	err = systemd.InstallUnit("f4f-manager", map[string]string{"Cwd": f4fManagerWd})
	if err != nil {
		return err
	}

	err = systemd.DaemonReload()
	if err != nil {
		return err
	}

	err = systemd.Enable("faasd-provider")
	if err != nil {
		return err
	}

	err = systemd.Enable("f4f-manager")
	if err != nil {
		return err
	}

	err = systemd.Start("faasd-provider")
	if err != nil {
		return err
	}

	err = systemd.Start("f4f-manager")
	if err != nil {
		return err
	}

	fmt.Println(`Check status with:
  sudo journalctl -u f4f-manager --lines 100 -f

Login with:
  sudo -E cat /var/lib/f4f-manager/secrets/basic-auth-password | forge4flow login -s`)

	return nil
}

func binExists(folder, name string) error {
	findPath := path.Join(folder, name)
	if _, err := os.Stat(findPath); err != nil {
		return fmt.Errorf("unable to stat %s, install this binary before continuing", findPath)
	}
	return nil
}
func ensureSecretsDir(folder string) error {
	if _, err := os.Stat(folder); err != nil {
		err = os.MkdirAll(folder, secretDirPermission)
		if err != nil {
			return err
		}
	}

	return nil
}
func ensureWorkingDir(folder string) error {
	if _, err := os.Stat(folder); err != nil {
		err = os.MkdirAll(folder, workingDirectoryPermission)
		if err != nil {
			return err
		}
	}

	return nil
}

func cp(source, destFolder string) error {
	file, err := os.Open(source)
	if err != nil {
		return err

	}
	defer file.Close()

	out, err := os.Create(path.Join(destFolder, source))
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, file)

	return err
}
