package main

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"log"
  "strings"
	"time"

	"github.com/hashicorp/terraform/communicator"
	"github.com/hashicorp/terraform/communicator/remote"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/terraform"
	"github.com/mitchellh/go-linereader"
)

var maxBackoffDelay = 10 * time.Second
var initialBackoffDelay = time.Second

func Provisioner() terraform.ResourceProvisioner {
	return &schema.Provisioner{
		Schema: map[string]*schema.Schema{
			"login": &schema.Schema{
				Type:          schema.TypeString,
				Required: true,
			},
      "password": &schema.Schema{
        Type:          schema.TypeString,
        Required: true,
      },
      "products": &schema.Schema{
          Required: true,
          ForceNew: true,
          Type: schema.TypeList,
          Elem: &schema.Resource{
          Schema: map[string]*schema.Schema{
            "identifier": {
              Type:     schema.TypeString,
              Optional: true,
            },
            "version": {
              Type:     schema.TypeString,
              Optional: true,
            },
            "arch": {
              Type:     schema.TypeString,
              Optional: true,
            },
            "regcode": {
              Type:     schema.TypeString,
              Optional: true,
            },
            "service_name": {
              Type:     schema.TypeString,
              //Required: true,
              Optional: true,
            },
            "service_url": {
              Type:     schema.TypeString,
              //Required: true,
              Optional: true,
            },
          },
        },
      },
		},

		ApplyFunc:    applyFn,
		//ValidateFunc: validateFn,
	}
}

// Apply executes the remote exec provisioner
func applyFn(ctx context.Context) error {
	connState := ctx.Value(schema.ProvRawStateKey).(*terraform.InstanceState)
	data := ctx.Value(schema.ProvConfigDataKey).(*schema.ResourceData)
	o := ctx.Value(schema.ProvOutputKey).(terraform.UIOutput)

	// Get a new communicator
	comm, err := communicator.New(connState)
	if err != nil {
		return err
	}

	// Collect the scripts
	scripts, err := collectScripts(data)
	if err != nil {
		return err
	}
	for _, s := range scripts {
		defer s.Close()
	}

	// Copy and execute each script
	if err := runScripts(ctx, o, comm, scripts); err != nil {
		return err
	}

	return nil
}

func collectScripts(d *schema.ResourceData) ([]io.ReadCloser, error) {
  var lines []string
  
  login    := d.Get("login").(string)
  password := d.Get("password").(string)
  
  lines = append(lines, fmt.Sprintf("echo 'username=%s' > /etc/zypp/credentials.d/SCCcredentials", login))
  lines = append(lines, fmt.Sprintf("echo 'password=%s' >> /etc/zypp/credentials.d/SCCcredentials", password))
  lines = append(lines, "chmod 600 /etc/zypp/credentials.d/SCCcredentials")
  
  products := d.Get("products").([]interface{})
  for _, product := range products {
    service_name :=  product.(map[string]interface {})["service_name"].(string)
    service_url  :=  product.(map[string]interface {})["service_url"].(string)
    
    lines = append(lines, fmt.Sprintf("echo 'username=%s' > /etc/zypp/credentials.d/%s", login, service_name))
    lines = append(lines, fmt.Sprintf("echo 'password=%s' >> /etc/zypp/credentials.d/%s", password, service_name))
    lines = append(lines, fmt.Sprintf("chmod 600 /etc/zypp/credentials.d/%s", service_name))
    lines = append(lines, fmt.Sprintf("zypper rs %s 2>/dev/null", service_name))
    lines = append(lines, fmt.Sprintf("zypper as %s %s", service_url, service_name))
  }

  scripts := []string{strings.Join(lines, "\n")}
  
	var r []io.ReadCloser
	for _, script := range scripts {
		r = append(r, ioutil.NopCloser(bytes.NewReader([]byte(script))))
	}

	return r, nil
}

/*

Everything below is copypasta from https://github.com/hashicorp/terraform/blob/e9e4ee494070950f31d17f29cae2d1facf5d564e/builtin/provisioners/remote-exec/resource_provisioner.go
, because those functions aren't exported.

*/

// runScripts is used to copy and execute a set of scripts
func runScripts(
	ctx context.Context,
	o terraform.UIOutput,
	comm communicator.Communicator,
	scripts []io.ReadCloser) error {

	retryCtx, cancel := context.WithTimeout(ctx, comm.Timeout())
	defer cancel()

	// Wait and retry until we establish the connection
	err := communicator.Retry(retryCtx, func() error {
		return comm.Connect(o)
	})
	if err != nil {
		return err
	}

	// Wait for the context to end and then disconnect
	go func() {
		<-ctx.Done()
		comm.Disconnect()
	}()

	for _, script := range scripts {
		var cmd *remote.Cmd

		outR, outW := io.Pipe()
		errR, errW := io.Pipe()
		defer outW.Close()
		defer errW.Close()

		go copyOutput(o, outR)
		go copyOutput(o, errR)

		remotePath := comm.ScriptPath()

		if err := comm.UploadScript(remotePath, script); err != nil {
			return fmt.Errorf("Failed to upload script: %v", err)
		}

		cmd = &remote.Cmd{
			Command: remotePath,
			Stdout:  outW,
			Stderr:  errW,
		}
		if err := comm.Start(cmd); err != nil {
			return fmt.Errorf("Error starting script: %v", err)
		}

		if err := cmd.Wait(); err != nil {
			return err
		}

		// Upload a blank follow up file in the same path to prevent residual
		// script contents from remaining on remote machine
		empty := bytes.NewReader([]byte(""))
		if err := comm.Upload(remotePath, empty); err != nil {
			// This feature is best-effort.
			log.Printf("[WARN] Failed to upload empty follow up script: %v", err)
		}
	}

	return nil
}

func copyOutput(
	o terraform.UIOutput, r io.Reader) {
	lr := linereader.New(r)
	for line := range lr.Ch {
		o.Output(line)
	}
}
