package main

type ParamsInit struct {
	Path string `arg:"" type:"path" default:"pantest.yml" help:"Path of the config file to create."`
}

type ParamsRun struct {
	ConfigFile string `short:"f" type:"existingfile" default:"pantest.yml" help:"Choose the configuration file."`
}

type ParamsSame struct {
	ParamsRun
	Rounds int    `arg:"" help:"Number of tests to run."`
	Exec1  string `arg:"" help:"Name of the first executable to test."`
	Exec2  string `arg:"" help:"Name of the second executable to test."`
}

type Cli struct {
	Init ParamsInit `cmd:"" help:"Create a new configuration file."`
	Run  ParamsRun  `cmd:"" help:"Run the tests."`
	Same ParamsSame `cmd:"" help:"Check the functional similarity of two programs with random tests."`
}

func (p *ParamsInit) Run() error {
	return initConfigFile(p.Path)
}

func (p *ParamsRun) Run() error {
	runCli(p.ConfigFile)
	return nil
}

func (p *ParamsSame) Run() error {
	runSame(p.ConfigFile, p.Rounds, p.Exec1, p.Exec2)
	return nil
}
