package ssh

import (
	"encoding/base64"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"os"
	"path/filepath"
	"time"

	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/agent"
	"golang.org/x/crypto/ssh/knownhosts"
	"ssh-tunnel/config"
)

func NewSSHClient(sv config.Server) (*ssh.Client, error) {
	server := sv.ServerAddress
	user := sv.User

	auths, err := getAuths(sv)

	if err != nil {
		return nil, fmt.Errorf("failed to get auths: %w", err)
	}

	timeout := time.Second * 5

	if sv.Timeout > 0 {
		timeout = time.Second * time.Duration(sv.Timeout)
	}

	var hostKeyCallback = getHostKeyCallback(sv)

	sshClient, err := ssh.Dial("tcp", server, &ssh.ClientConfig{
		User:            user,
		Auth:            auths,
		HostKeyCallback: hostKeyCallback,
		Timeout:         timeout,
	})

	if err != nil {
		return nil, fmt.Errorf("failed to dial: %w", err)
	}

	log.Printf("connected to %s", server)
	return sshClient, nil
}

func getHostKeyCallback(sv config.Server) ssh.HostKeyCallback {
	if sv.HotkeyCheck {
		return func(host string, remote net.Addr, pubKey ssh.PublicKey) error {
			var keyErr = new(knownhosts.KeyError)
			kh := checkKnownHosts()
			hErr := kh(host, remote, pubKey)
			// Reference: https://blog.golang.org/go1.13-errors
			// To understand what errors.As is.
			if errors.As(hErr, &keyErr) && len(keyErr.Want) > 0 {
				// Reference: https://www.godoc.org/golang.org/x/crypto/ssh/knownhosts#KeyError
				// if keyErr.Want slice is empty then host is unknown, if keyErr.Want is not empty
				// and if host is known then there is key mismatch the connection is then rejected.
				log.Printf("WARNING: %v is not a key of %s, either a MiTM attack or %s has reconfigured the host pub key.", hostKeyString(pubKey), host, host)
				return hErr
			} else if errors.As(hErr, &keyErr) && len(keyErr.Want) == 0 {
				// host key not found in known_hosts then give a warning and continue to connect.
				log.Printf("WARNING: %s is not trusted, adding this key: %q to known_hosts file.", host, hostKeyString(pubKey))
				return addHostKey(host, remote, pubKey)
			}
			log.Printf("Pub key exists for %s.", host)
			return nil
		}
	}
	return ssh.InsecureIgnoreHostKey()
}

func getAuths(sv config.Server) ([]ssh.AuthMethod, error) {
	auths := make([]ssh.AuthMethod, 0, 2)

	if sv.IdentityFile != "" {
		content, err := ioutil.ReadFile(sv.IdentityFile)
		if err != nil {
			return nil, fmt.Errorf("failed to read identity file: %w", err)
		}

		key, err := ssh.ParsePrivateKey(content)
		if err != nil {
			return nil, fmt.Errorf("failed to parse private key: %w", err)
		}
		auths = append(auths, ssh.PublicKeys(key))
	}

	return append(auths, sshAgent()), nil
}

func sshAgent() ssh.AuthMethod {
	if sshAgent, err := net.Dial("unix", os.Getenv("SSH_AUTH_SOCK")); err == nil {
		return ssh.PublicKeysCallback(agent.NewClient(sshAgent).Signers)
	}
	return nil
}

// create human-readable SSH-key strings
func hostKeyString(k ssh.PublicKey) string {
	return k.Type() + " " + base64.StdEncoding.EncodeToString(k.Marshal()) // e.g. "ecdsa-sha2-nistp256 AAAAE2VjZHNhLXNoYTItbmlzdHAyNTY...."
}

func checkKnownHosts() ssh.HostKeyCallback {
	createKnownHosts()
	kh, e := knownhosts.New(filepath.Join(os.Getenv("HOME"), ".ssh", "known_hosts"))
	errCallBack(e)
	return kh
}

func addHostKey(_ string, remote net.Addr, pubKey ssh.PublicKey) error {
	// add host key if host is not found in known_hosts, error object is return, if nil then connection proceeds,
	// if not nil then connection stops.
	khFilePath := filepath.Join(os.Getenv("HOME"), ".ssh", "known_hosts")

	f, fErr := os.OpenFile(khFilePath, os.O_APPEND|os.O_WRONLY, 0600)
	if fErr != nil {
		return fErr
	}
	defer f.Close()

	knownHosts := knownhosts.Normalize(remote.String())
	_, fileErr := f.WriteString(knownhosts.Line([]string{knownHosts}, pubKey))
	return fileErr
}

func createKnownHosts() {
	f, fErr := os.OpenFile(filepath.Join(os.Getenv("HOME"), ".ssh", "known_hosts"), os.O_CREATE, 0600)
	if fErr != nil {
		log.Fatal(fErr)
	}
	f.Close()
}

func errCallBack(e error) {
	if e != nil {
		log.Fatal(e)
	}
}
