package cosmosgen_test

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/ignite-hq/cli/ignite/chainconfig"
	"github.com/ignite-hq/cli/ignite/pkg/cmdrunner/step"
	"github.com/ignite-hq/cli/ignite/pkg/xurl"
	envtest "github.com/ignite-hq/cli/integration"
	"github.com/stretchr/testify/require"
)

func TestCustomModule(t *testing.T) {
	var (
		env  = envtest.New(t)
		path = env.Scaffold("chain", "--no-module")
		host = env.RandomizeServerPorts(path, "")
	)

	queryAPI, err := xurl.HTTP(host.API)
	require.NoError(t, err)

	txAPI, err := xurl.TCP(host.RPC)
	require.NoError(t, err)

	// Accounts to be included in the genesis
	accounts := []chainconfig.Account{
		{
			Name:    "account1",
			Address: "cosmos1j8hw8283hj80hhq8urxaj40syrzqp77dt8qwhm",
			Mnemonic: fmt.Sprint(
				"toe mail light plug pact length excess predict real artwork laundry when ",
				"steel online adapt clutch debate vehicle dash alter rifle virtual season almost",
			),
			Coins: []string{"10000token", "10000stake"},
		},
	}

	env.UpdateConfig(path, "", func(cfg *chainconfig.Config) error {
		cfg.Accounts = append(cfg.Accounts, accounts...)
		return nil
	})

	env.Must(env.Exec("create a module",
		step.NewSteps(step.New(
			step.Exec(envtest.IgniteApp, "s", "module", "disco", "--require-registration", "--yes"),
			step.Workdir(path),
		)),
	))

	env.Must(env.Exec("create a list type",
		step.NewSteps(step.New(
			step.Exec(envtest.IgniteApp, "s", "list", "entry", "name", "--module", "disco", "--yes"),
			step.Workdir(path),
		)),
	))

	env.Must(env.Exec("generate vuex store", step.NewSteps(
		step.New(
			step.Exec(envtest.IgniteApp, "g", "vuex", "--proto-all-modules", "--yes"),
			step.Workdir(path),
		),
	)))

	ctx, cancel := context.WithTimeout(env.Ctx(), envtest.ServeTimeout)
	defer cancel()

	go func() {
		env.Serve("should serve app", path, "", "", envtest.ExecCtx(ctx))
	}()

	// Wait for the server to be up before running the client tests
	err = env.IsAppServed(ctx, host)
	require.NoError(t, err)

	testAccounts, err := json.Marshal(accounts)
	require.NoError(t, err)

	env.Must(env.RunClientTests(
		path,
		envtest.ClientTestFile("custom_module_test.ts"),
		envtest.ClientEnv(map[string]string{
			"TEST_QUERY_API": queryAPI,
			"TEST_TX_API":    txAPI,
			"TEST_ACCOUNTS":  string(testAccounts),
		}),
	))
}
