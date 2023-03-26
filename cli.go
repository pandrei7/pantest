package main

type ParamsInit struct {
	Path string `arg:"" type:"path" default:"pantest.yml" help:"Path of the config file to create."`
}

type ParamsRun struct {
	ConfigFile string `short:"f" type:"existingfile" default:"pantest.yml" help:"Choose the configuration file."`
}

type Cli struct {
	Init ParamsInit `cmd:"" help:"Create a new configuration file."`
	Run  ParamsRun  `cmd:"" help:"Run the tests."`
}

func (p *ParamsInit) Run() error {
	return initConfigFile(p.Path)
}

func (p *ParamsRun) Run() error {
	runCli(p.ConfigFile)
	return nil
}
