package common

import (
	"fmt"
	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/packer"
)

type StepForwardPortOverSSH struct {
	// If NoProxy, don't do proxying (forwarding)
	NoProxy    bool
	RemotePort func(state multistep.StateBag) (uint, error)
	RemoteDest func(state multistep.StateBag) (string, error)

	HostPortMin uint
	HostPortMax uint

	ResultKey string
}

func (self *StepForwardPortOverSSH) Run(state multistep.StateBag) multistep.StepAction {

	config := state.Get("commonconfig").(CommonConfig)
	ui := state.Get("ui").(packer.Ui)
	if self.NoProxy {
		ui.Say(fmt.Sprintf("Not using SSH port forwarding"))
		state.Put(self.ResultKey, uint(0))
		return multistep.ActionContinue
	}

	// Find a free local port:

	l, sshHostPort := FindPort(self.HostPortMin, self.HostPortMax)

	if l == nil || sshHostPort == 0 {
		ui.Error("Error: unable to find free host port. Try providing a larger range [host_port_min, host_port_max]")
		return multistep.ActionHalt
	}

	ui.Say(fmt.Sprintf("Creating a local port forward over SSH on local port %d", sshHostPort))

	hostAddress, _ := state.Get("ssh_address").(string)
	remotePort, _ := self.RemotePort(state)
	remoteDest, _ := self.RemoteDest(state)

	go ssh_port_forward(l, remotePort, remoteDest, hostAddress, config.Username, config.Password)
	ui.Say(fmt.Sprintf("Port forward setup. %d ---> %s:%d on %s", sshHostPort, remoteDest, remotePort, hostAddress))

	// Provide the local port to future steps.
	state.Put(self.ResultKey, sshHostPort)

	return multistep.ActionContinue
}

func (self *StepForwardPortOverSSH) Cleanup(state multistep.StateBag) {}
