package xva

import (
	"fmt"
	"os"

	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/packer"
	xsclient "github.com/xenserver/go-xenserver-client"
	xscommon "github.com/xenserver/packer-builder-xenserver/builder/xenserver/common"
)

type stepImportInstance struct {
	instance *xsclient.VM
	vdi      *xsclient.VDI
}

func (self *stepImportInstance) Run(state multistep.StateBag) multistep.StepAction {
	config := state.Get("config").(config)

	if state.Get("instance_uuid") != nil {
		return multistep.ActionContinue
	}

	ui := state.Get("ui").(packer.Ui)
	client := state.Get("client").(xsclient.XenAPIClient)

	ui.Say("Step: Import Instance")

	if config.SourcePath == "" {
		ui.Error(fmt.Sprintf("Failed to instantiate \"source_template\": \"%s\" and \"source_path\" is empty. Aborting.", config.SourceTemplate))
		return multistep.ActionHalt
	}

	// find the SR
	sr, err := config.GetSR(client)
	if err != nil {
		ui.Error(fmt.Sprintf("Unable to get SR: %s", err.Error()))
		return multistep.ActionHalt
	}

	// Open the file for reading (NB: httpUpload closes the file for us)
	fh, err := os.Open(config.SourcePath)
	if err != nil {
		ui.Error(fmt.Sprintf("Unable to open XVA '%s': %s", config.SourcePath, err.Error()))
		return multistep.ActionHalt
	}

	result, err := xscommon.HTTPUpload(fmt.Sprintf("https://%s/import?session_id=%s&sr_id=%s",
		client.Host,
		client.Session.(string),
		sr.Ref,
	), fh, state)
	if err != nil {
		ui.Error(fmt.Sprintf("Unable to upload VDI: %s", err.Error()))
		return multistep.ActionHalt
	}
	if result == nil {
		ui.Error("XAPI did not reply with an instance reference")
		return multistep.ActionHalt
	}

	instance := xsclient.VM(*result)

	instanceId, err := instance.GetUuid()
	if err != nil {
		ui.Error(fmt.Sprintf("Unable to get VM UUID: %s", err.Error()))
		return multistep.ActionHalt
	}
	state.Put("instance_uuid", instanceId)

	instance.SetDescription(config.VMDescription)
	if err != nil {
		ui.Error(fmt.Sprintf("Error setting VM description: %s", err.Error()))
		return multistep.ActionHalt
	}

	ui.Say(fmt.Sprintf("Imported instance '%s'", instanceId))

	return multistep.ActionContinue
}

func (self *stepImportInstance) Cleanup(state multistep.StateBag) {
	/*
		config := state.Get("config").(config)
		if config.ShouldKeepVM(state) {
			return
		}

		ui := state.Get("ui").(packer.Ui)

		if self.instance != nil {
			ui.Say("Destroying VM")
			_ = self.instance.HardShutdown() // redundant, just in case
			err := self.instance.Destroy()
			if err != nil {
				ui.Error(err.Error())
			}
		}

		if self.vdi != nil {
			ui.Say("Destroying VDI")
			err := self.vdi.Destroy()
			if err != nil {
				ui.Error(err.Error())
			}
		}
	*/
}
