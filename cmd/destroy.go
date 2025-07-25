// Copyright 2020 Nokia
// Licensed under the BSD 3-Clause License.
// SPDX-License-Identifier: BSD-3-Clause

package cmd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/charmbracelet/log"
	"github.com/spf13/cobra"
	"github.com/srl-labs/containerlab/clab"
	"github.com/srl-labs/containerlab/cmd/common"
	"github.com/srl-labs/containerlab/labels"
	"github.com/srl-labs/containerlab/links"
	"github.com/srl-labs/containerlab/runtime"
	"github.com/srl-labs/containerlab/types"
	"github.com/srl-labs/containerlab/utils"
)

var (
	all         bool
	cleanup     bool
	keepMgmtNet bool
)

// destroyCmd represents the destroy command.
var destroyCmd = &cobra.Command{
	Use:     "destroy",
	Short:   "destroy a lab",
	Long:    "destroy a lab based defined by means of the topology definition file\nreference: https://containerlab.dev/cmd/destroy/",
	Aliases: []string{"des"},
	PreRunE: common.CheckAndGetRootPrivs,
	RunE:    destroyFn,
}

func init() {
	RootCmd.AddCommand(destroyCmd)
	destroyCmd.Flags().BoolVarP(&cleanup, "cleanup", "c", false,
		"delete lab directory. Cannot be used with node-filter")
	destroyCmd.Flags().BoolVarP(&common.Graceful, "graceful", "", false,
		"attempt to stop containers before removing")
	destroyCmd.Flags().BoolVarP(&all, "all", "a", false, "destroy all containerlab labs")
	destroyCmd.Flags().UintVarP(&maxWorkers, "max-workers", "", 0,
		"limit the maximum number of workers deleting nodes")
	destroyCmd.Flags().BoolVarP(&keepMgmtNet, "keep-mgmt-net", "", false, "do not remove the management network")
	destroyCmd.Flags().StringSliceVarP(&common.NodeFilter, "node-filter", "", []string{},
		"comma separated list of nodes to include")
}

func destroyFn(_ *cobra.Command, _ []string) error {
	// cleanup doesn't make sense with node-filter
	if len(common.NodeFilter) != 0 && cleanup {
		return fmt.Errorf("cleanup cannot be used with node-filter")
	}

	var err error
	var labs []*clab.CLab
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// topo will hold the reference to the topology file
	// as the key and the respective lab directory as the referenced value
	topos := map[string]string{}

	switch {
	case !all:
		log.Debug("topology file", "file", common.Topo)
		cnts, err := listContainers(ctx, common.Topo)
		if err != nil {
			return err
		}

		if len(cnts) == 0 {
			log.Info("No containerlab containers found")
			if cleanup {
				// do our best to find a labdir
				var labDirs []string
				if common.Topo != "" {
					topoDir := filepath.Dir(common.Topo)
					log.Debug("Looking for lab directory next to topology file", "path", topoDir)
					labDirs, _ = filepath.Glob(filepath.Join(topoDir, "clab-*"))
				} else if len(labDirs) == 0 {
					// Look in the current directory
					log.Debug("Looking for lab directory in current directory")
					labDirs, _ = filepath.Glob("clab-*")
				}
				if len(labDirs) != 0 {
					// we only really care about the first found
					log.Info("Removing lab directory", "path", labDirs[0])
					if err := os.RemoveAll(labDirs[0]); err != nil {
						log.Errorf("error deleting lab directory: %v", err)
					}
				}
			}
			return nil
		}

		topos[cnts[0].Labels[labels.TopoFile]] =
			filepath.Dir(cnts[0].Labels[labels.NodeLabDir])

	case all:
		containers, err := listContainers(ctx, common.Topo)
		if err != nil {
			return err
		}

		if len(containers) == 0 {
			return fmt.Errorf("no containerlab labs found")
		}
		// get unique topo files from all labs
		for i := range containers {
			topos[containers[i].Labels[labels.TopoFile]] =
				filepath.Dir(containers[i].Labels[labels.NodeLabDir])
		}
	}

	log.Debugf("We got the following topos struct for destroy: %+v", topos)
	for topo, labdir := range topos {
		opts := []clab.ClabOption{
			clab.WithTimeout(common.Timeout),
			clab.WithTopoPath(topo, common.VarsFile),
			clab.WithNodeFilter(common.NodeFilter),
			clab.WithRuntime(common.Runtime,
				&runtime.RuntimeConfig{
					Debug:            common.Debug,
					Timeout:          common.Timeout,
					GracefulShutdown: common.Graceful,
				},
			),
			clab.WithDebug(common.Debug),
			// during destroy we don't want to check bind paths
			// as it is irrelevant for this command.
			clab.WithSkippedBindsPathsCheck(),
		}

		if keepMgmtNet {
			opts = append(opts, clab.WithKeepMgmtNet())
		}

		log.Debugf("going through extracted topos for destroy, got a topo file %v and generated opts list %+v", topo, opts)
		nc, err := clab.NewContainerLab(opts...)
		if err != nil {
			return err
		}

		// check if labdir exists and is a directory
		if labdir != "" && utils.FileOrDirExists(labdir) {
			// adjust the labdir. Usually we take the PWD. but now on destroy time,
			// we might be in a different Dir.
			err = nc.TopoPaths.SetLabDir(labdir)
			if err != nil {
				return err
			}
		}

		err = links.SetMgmtNetUnderlyingBridge(nc.Config.Mgmt.Bridge)
		if err != nil {
			return err
		}

		// create management network or use existing one
		// we call this to populate the nc.cfg.mgmt.bridge variable
		// which is needed for the removal of the iptables rules
		if err = nc.CreateNetwork(ctx); err != nil {
			return err
		}

		err = nc.ResolveLinks()
		if err != nil {
			return err
		}

		labs = append(labs, nc)
	}

	var errs []error
	for _, clab := range labs {
		err = destroyLab(ctx, clab)
		if err != nil {
			log.Errorf("Error occurred during the %s lab deletion: %v", clab.Config.Name, err)
			errs = append(errs, err)
		}

		if cleanup {
			err = os.RemoveAll(clab.TopoPaths.TopologyLabDir())
			if err != nil {
				log.Errorf("error deleting lab directory: %v", err)
			}
		}
	}

	if len(errs) != 0 {
		return fmt.Errorf("error(s) occurred during the deletion. Check log messages")
	}

	return nil
}

func destroyLab(ctx context.Context, c *clab.CLab) (err error) {
	return c.Destroy(ctx, maxWorkers, keepMgmtNet)
}

// listContainers lists containers belonging to a certain topo if topo file path is specified
// otherwise lists all containerlab containers.
func listContainers(ctx context.Context, topo string) ([]runtime.GenericContainer, error) {
	runtimeConfig := &runtime.RuntimeConfig{
		Debug:            common.Debug,
		Timeout:          common.Timeout,
		GracefulShutdown: common.Graceful,
	}

	opts := []clab.ClabOption{
		clab.WithRuntime(common.Runtime, runtimeConfig),
		clab.WithTimeout(common.Timeout),
		// when listing containers we don't care if binds are accurate
		// since this function is used in the destroy process
		clab.WithSkippedBindsPathsCheck(),
	}

	// filter to list all containerlab containers
	// it is overwritten if topo file is provided
	filter := []*types.GenericFilter{{
		FilterType: "label",
		Field:      labels.Containerlab,
		Operator:   "exists",
	}}

	// when topo file is provided, filter containers by lab name
	if topo != "" {
		opts = append(opts, clab.WithTopoPath(topo, common.VarsFile))
	}

	c, err := clab.NewContainerLab(opts...)
	if err != nil {
		return nil, err
	}

	if topo != "" {
		filter = []*types.GenericFilter{{
			FilterType: "label",
			Field:      labels.Containerlab,
			Operator:   "=",
			Match:      c.Config.Name,
		}}
	}

	containers, err := c.ListContainers(ctx, filter)
	if err != nil {
		return nil, err
	}

	return containers, nil
}
