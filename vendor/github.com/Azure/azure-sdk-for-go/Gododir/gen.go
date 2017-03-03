package main

// To run this package...
// go run gen.go -- --sdk 3.14.16

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"

	do "gopkg.in/godo.v2"
)

type service struct {
	Name        string
	Fullname    string
	Namespace   string
	Packages    []string
	TaskName    string
	Version     string
	Input       string
	Output      string
	Swagger     string
	SubServices []service
}

type mapping struct {
	Plane       string
	InputPrefix string
	Services    []service
}

var (
	gopath          = os.Getenv("GOPATH")
	sdkVersion      string
	autorestDir     string
	swaggersDir     string
	deps            do.S
	services        = []*service{}
	servicesMapping = []mapping{
		{
			Plane:       "arm",
			InputPrefix: "arm-",
			Services: []service{
				{
					Name:    "analysisservices",
					Version: "2016-05-16",
				},
				{
					Name:    "authorization",
					Version: "2015-07-01",
				},
				{
					Name:    "batch",
					Version: "2015-12-01",
					Swagger: "BatchManagement",
				},
				{
					Name:    "cdn",
					Version: "2016-10-02",
				},
				{
					Name:    "cognitiveservices",
					Version: "2016-02-01-preview",
				},
				{
					Name:    "commerce",
					Version: "2015-06-01-preview",
				},
				{
					Name:    "compute",
					Version: "2016-03-30",
				},
				{
					Name:    "containerservice",
					Version: "2016-09-30",
					Swagger: "containerService",
					Input:   "compute",
				},
				{
					Name:    "containerregistry",
					Version: "2016-06-27-preview",
				},
				{
					Name: "datalake-analytics",
					SubServices: []service{
						{
							Name:    "account",
							Version: "2016-11-01",
						},
					},
				},
				{
					Name: "datalake-store",
					SubServices: []service{
						{
							Name:    "account",
							Version: "2016-11-01",
						},
					},
				},
				{
					Name:    "devtestlabs",
					Version: "2016-05-15",
					Swagger: "DTL",
				},
				{
					Name:    "dns",
					Version: "2016-04-01",
				},
				{
					Name:    "documentdb",
					Version: "2015-04-08",
				},
				{
					Name:    "eventhub",
					Version: "2015-08-01",
					Swagger: "EventHub",
				},
				// {
				// 	Name:    "graphrbac",
				// 	Version: "1.6",
				// 	// Composite swagger
				// },
				// {
				// 	Name:    "insights",
				// 	// Composite swagger
				// },
				{
					Name:    "intune",
					Version: "2015-01-14-preview",
				},
				{
					Name:    "iothub",
					Version: "2016-02-03",
				},
				{
					Name:    "keyvault",
					Version: "2015-06-01",
				},
				{
					Name:    "logic",
					Version: "2016-06-01",
					// composite swagger
				},
				{
					Name: "machinelearning",
					SubServices: []service{
						{
							Name:    "webservices",
							Version: "2016-05-01-preview",
							Input:   "machinelearning",
						},
						{
							Name:    "commitmentplans",
							Version: "2016-05-01-preview",
							Swagger: "commitmentPlans",
							Input:   "machinelearning",
						},
					},
				},
				{
					Name:    "mediaservices",
					Version: "2015-10-01",
					Swagger: "media",
				},
				{
					Name:    "mobileengagement",
					Version: "2014-12-01",
					Swagger: "mobile-engagement",
				},
				{
					Name:    "network",
					Version: "2016-09-01",
				},
				{
					Name:    "notificationhubs",
					Version: "2016-03-01",
				},
				{
					Name:    "powerbiembedded",
					Version: "2016-01-29",
				},
				{
					Name:    "recoveryservices",
					Version: "2016-06-01",
				},
				// {
				// 	Name:    "recoveryservicesbackup",
				// 	Version: "2016-06-01",
				// composite swagger
				// },
				{
					Name:    "redis",
					Version: "2016-04-01",
				},
				{
					Name: "resources",
					SubServices: []service{
						{
							Name:    "features",
							Version: "2015-12-01",
						},
						{
							Name:    "links",
							Version: "2016-09-01",
						},
						{
							Name:    "locks",
							Version: "2016-09-01",
						},
						{
							Name:    "policy",
							Version: "2016-04-01",
						},
						{
							Name:    "resources",
							Version: "2016-09-01",
						},
						{
							Name:    "subscriptions",
							Version: "2016-06-01",
						},
					},
				},
				{
					Name:    "scheduler",
					Version: "2016-03-01",
				},
				{
					Name:    "search",
					Version: "2015-08-19",
				},
				{
					Name:    "servermanagement",
					Version: "2016-07-01-preview",
				},
				{
					Name:    "servicebus",
					Version: "2015-08-01",
				},
				{
					Name:    "sql",
					Version: "2014-04-01",
					Swagger: "sql.core",
				},
				{
					Name:    "storage",
					Version: "2016-01-01",
				},
				{
					Name:    "trafficmanager",
					Version: "2015-11-01",
				},
				{
					Name:    "web",
					Version: "2015-08-01",
					Swagger: "service",
					// enormous composite swagger
				},
			},
		},
		{
		// Plane:       "dataplane",
		// InputPrefix: "",
		// Services: []Service{
		// 	{
		// 		Name:    "batch",
		// 		Version: "2016-07-01.3.1",
		// 		Swagger: "BatchService",
		// 	},
		//     {
		//         Name: "insights",
		//         // composite swagger
		//     },
		//     {
		//         Name: "keyvault",
		//         Version: "2015-06-01",
		//     },
		//     {
		//         Name: "search",'
		//         Version: "2015-02-28"
		//         // There are 2 files, but no composite swagger...
		//     },
		//     {
		//         Name: "servicefabric",
		//         Version: "2016-01-28",
		//     },
		// },
		},
		{
			Plane:       "",
			InputPrefix: "arm-",
			Services: []service{
				{
					Name: "datalake-store",
					SubServices: []service{
						{
							Name:    "filesystem",
							Version: "2016-11-01",
						},
					},
				},
				// {
				// 	Name: "datalake-analytics",
				// 	SubServices: []Service{
				// 		{
				// 			Name:    "catalog",
				// 			Version: "2016-06-01-preview",
				// 		},
				// 		{
				// 			Name:    "job",
				// 			Version: "2016-03-20-preview",
				// 		},
				// 	},
				// },
			},
		},
	}
)

func main() {
	for _, swaggerGroup := range servicesMapping {
		swg := swaggerGroup
		for _, service := range swg.Services {
			s := service
			initAndAddService(&s, swg.InputPrefix, swg.Plane)
		}
	}
	do.Godo(tasks)
}

func initAndAddService(service *service, inputPrefix, plane string) {
	if service.Swagger == "" {
		service.Swagger = service.Name
	}
	packages := append(service.Packages, service.Name)
	service.TaskName = fmt.Sprintf("%s>%s", plane, strings.Join(packages, ">"))
	service.Fullname = fmt.Sprintf("%s/%s", plane, strings.Join(packages, "/"))
	if service.Input == "" {
		service.Input = fmt.Sprintf("%s%s/%s/swagger/%s", inputPrefix, strings.Join(packages, "/"), service.Version, service.Swagger)
	} else {
		service.Input = fmt.Sprintf("%s%s/%s/swagger/%s", inputPrefix, service.Input, service.Version, service.Swagger)
	}
	service.Namespace = fmt.Sprintf("github.com/Azure/azure-sdk-for-go/%s", service.Fullname)
	service.Output = fmt.Sprintf("%s/src/%s", gopath, service.Namespace)

	if service.SubServices != nil {
		for _, subs := range service.SubServices {
			ss := subs
			ss.Packages = append(ss.Packages, service.Name)
			initAndAddService(&ss, inputPrefix, plane)
		}
	} else {
		services = append(services, service)
		deps = append(deps, service.TaskName)
	}
}

func tasks(p *do.Project) {
	p.Task("default", do.S{"setvars", "generate:all"}, nil)
	p.Task("setvars", nil, setVars)
	p.Use("generate", generateTasks)
	p.Use("gofmt", formatTasks)
	p.Use("gobuild", buildTasks)
	p.Use("golint", lintTasks)
	p.Use("govet", vetTasks)
	p.Use("delete", deleteTasks)
}

func setVars(c *do.Context) {
	if gopath == "" {
		panic("Gopath not set\n")
	}

	sdkVersion = c.Args.MustString("s", "sdk", "version")
	autorestDir = c.Args.MayString("C:", "a", "ar", "autorest")
	swaggersDir = c.Args.MayString("C:", "w", "sw", "swagger")
}

func generateTasks(p *do.Project) {
	addTasks(generate, p)
}

func generate(service *service) {
	fmt.Printf("Generating %s...\n\n", service.Fullname)
	delete(service)

	autorest := exec.Command(fmt.Sprintf("%s/autorest/src/core/AutoRest/bin/Debug/net451/win7-x64/autorest", autorestDir),
		"-Input", fmt.Sprintf("%s/azure-rest-api-specs/%s.json", swaggersDir, service.Input),
		"-CodeGenerator", "Go",
		"-Header", "MICROSOFT_APACHE",
		"-Namespace", service.Namespace,
		"-OutputDirectory", service.Output,
		"-Modeler", "Swagger",
		"-pv", sdkVersion)
	err := runner(autorest)
	if err != nil {
		panic(fmt.Errorf("Autorest error: %s", err))
	}

	format(service)
	build(service)
	lint(service)
	vet(service)
}

func deleteTasks(p *do.Project) {
	addTasks(format, p)
}

func delete(service *service) {
	fmt.Printf("Deleting %s...\n\n", service.Fullname)
	err := os.RemoveAll(service.Output)
	if err != nil {
		panic(fmt.Sprintf("Error deleting %s : %s\n", service.Output, err))
	}
}

func formatTasks(p *do.Project) {
	addTasks(format, p)
}

func format(service *service) {
	fmt.Printf("Formatting %s...\n\n", service.Fullname)
	gofmt := exec.Command("gofmt", "-w", service.Output)
	err := runner(gofmt)
	if err != nil {
		panic(fmt.Errorf("gofmt error: %s", err))
	}
}

func buildTasks(p *do.Project) {
	addTasks(build, p)
}

func build(service *service) {
	fmt.Printf("Building %s...\n\n", service.Fullname)
	gobuild := exec.Command("go", "build", service.Namespace)
	err := runner(gobuild)
	if err != nil {
		panic(fmt.Errorf("go build error: %s", err))
	}
}

func lintTasks(p *do.Project) {
	addTasks(lint, p)
}

func lint(service *service) {
	fmt.Printf("Linting %s...\n\n", service.Fullname)
	golint := exec.Command(fmt.Sprintf("%s/bin/golint", gopath), service.Namespace)
	err := runner(golint)
	if err != nil {
		panic(fmt.Errorf("golint error: %s", err))
	}
}

func vetTasks(p *do.Project) {
	addTasks(vet, p)
}

func vet(service *service) {
	fmt.Printf("Vetting %s...\n\n", service.Fullname)
	govet := exec.Command("go", "vet", service.Namespace)
	err := runner(govet)
	if err != nil {
		panic(fmt.Errorf("go vet error: %s", err))
	}
}

func addTasks(fn func(*service), p *do.Project) {
	for _, service := range services {
		s := service
		p.Task(s.TaskName, nil, func(c *do.Context) {
			fn(s)
		})
	}
	p.Task("all", deps, nil)
}

func runner(cmd *exec.Cmd) error {
	var stdout, stderr bytes.Buffer
	cmd.Stdout, cmd.Stderr = &stdout, &stderr
	err := cmd.Run()
	if stdout.Len() > 0 {
		fmt.Println(stdout.String())
	}
	if stderr.Len() > 0 {
		fmt.Println(stderr.String())
	}
	return err
}
