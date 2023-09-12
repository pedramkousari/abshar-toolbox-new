package toolbox

import (
	"context"
	"fmt"
	"log"
	"os"
	"path"

	"github.com/pedramkousari/abshar-toolbox-new/config"
	"github.com/pedramkousari/abshar-toolbox-new/contracts"
	"github.com/pedramkousari/abshar-toolbox-new/pkg/logger"
	"github.com/pedramkousari/abshar-toolbox-new/utils"
)

var excludePath = []string{}

var appendPath = []string{}

func NewGenerator(cnf config.Config, tag1 string, tag2 string, loading contracts.Loader) *toolbox {

	if tag1 == "" {
		log.Fatal("tag 1 not initialized")
	}

	if tag2 == "" {
		log.Fatal("tag 2 not initialized")
	}

	return &toolbox{
		tempDir:     cnf.TempDir + "/toolbox",
		outFile:     cnf.TempDir + "/builds/toolbox.tar.gz",
		dir:         path.Join(cnf.DockerComposeDir, "services/update-toolbox"),
		tag1:        tag1,
		tag2:        tag2,
		percent:     0,
		loading:     loading,
		serviceName: "toolbox",
	}
}

func (t *toolbox) GetFilePath() string {
	return t.outFile
}

func (t *toolbox) Generate(ctx context.Context) error {

	completeSignal := make(chan error)
	go func() {
		defer close(completeSignal)
		if err := t.runGenerate(ctx); err != nil {
			completeSignal <- err
		}
	}()

	select {
	case err, ok := <-completeSignal:
		if !ok {
			logger.Info(fmt.Sprintf("Service Generate Package %s Completed", t.serviceName))
			return nil
		}

		if err != nil {
			return fmt.Errorf("Service Generate Package %s is failed: %v", t.serviceName, err)
		}

		return nil

	case <-ctx.Done():
		logger.Info(fmt.Sprintf("%s Generate Package Canceled", t.serviceName))
		return ctx.Err()
	}
}

func (t *toolbox) runGenerate(ctx context.Context) error {
	var err error

	err = os.Mkdir(t.tempDir, 0755)
	if err != nil {
		return err
	}

	t.exec(ctx, 5, "Removed Tag", func() error {
		utils.RemoveTag(t.dir, t.tag2)
		return nil
	})

	err = t.exec(ctx, 10, "Fetch Sync With Git Server", func() error {
		return utils.Fetch(t.dir)
	})
	if err != nil {
		return fmt.Errorf("Cannot Fetch: %v", err)
	}

	err = t.exec(ctx, 20, "Get Diff Code", func() error {
		return utils.GetDiff(t.dir, t.tag1, t.tag2, excludePath, appendPath, t.serviceName)
	})
	if err != nil {
		return fmt.Errorf("Cannot Get Diff: %v", err)
	}

	err = t.exec(ctx, 30, "Create Tar File", func() error {
		return utils.CreateTarFile(t.dir, t.tempDir)
	})
	if err != nil {
		return fmt.Errorf("Cannot Create Tar File: %v", err)
	}

	err = t.exec(ctx, 90, "Copy Tar File To Temp Directory", func() error {
		return os.Rename(t.dir+"/patch.tar", t.tempDir+"/patch.tar")
	})
	if err != nil {
		return fmt.Errorf("Cannot Copy Tar File To Temp Directory: %v", err)
	}

	err = t.exec(ctx, 98, "Gzip Tar File", func() error {
		return utils.GzipTarFile(t.tempDir)
	})
	if err != nil {
		return fmt.Errorf("Cannot Gzip Tar File: %v", err)
	}

	err = t.exec(ctx, 100, "Gzip Tar File", func() error {
		return os.Rename(t.tempDir+"/patch.tar.gz", t.GetFilePath())
	})
	if err != nil {
		return fmt.Errorf("Cannot Gzip Tar File: %v", err)
	}

	return nil
}
