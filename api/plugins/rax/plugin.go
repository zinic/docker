package rax

import (
    "os"
    "fmt"
    "time"
    "strings"

    "github.com/dotcloud/docker/api/client"
    "github.com/dotcloud/docker/utils"

    "github.com/nu7hatch/gouuid"
    "github.com/rackspace/gophercloud"
)

type RaxPlugin struct {
    name string
}

var (
    nilOptions = gophercloud.AuthOptions{}

    // ErrNoAuthUrl errors occur when the value of the OS_AUTH_URL environment variable cannot be determined.
    ErrNoAuthUrl = fmt.Errorf("Environment variable OS_AUTH_URL needs to be set.")

    // ErrNoUsername errors occur when the value of the OS_USERNAME environment variable cannot be determined.
    ErrNoUsername = fmt.Errorf("Environment variable OS_USERNAME needs to be set.")

    // ErrNoPassword errors occur when the value of the OS_PASSWORD environment variable cannot be determined.
    ErrNoPassword = fmt.Errorf("Environment variable OS_PASSWORD or OS_API_KEY needs to be set.")
)

var Init = func() (client.CliPlugin, error) {
    cp := &RaxPlugin{
        name: "rax",
    }

    return cp, nil
}

func init() {
    client.RegisterPlugin("rax", Init)
}

func getAuthOptions() (string, gophercloud.AuthOptions, error) {
    provider := os.Getenv("OS_AUTH_URL")
    username := os.Getenv("OS_USERNAME")
    apiKey := os.Getenv("OS_API_KEY")
    password := os.Getenv("OS_PASSWORD")
    tenantId := os.Getenv("OS_TENANT_ID")
    tenantName := os.Getenv("OS_TENANT_NAME")

    if provider == "" {
        return "", nilOptions, ErrNoAuthUrl
    }

    if username == "" {
        return "", nilOptions, ErrNoUsername
    }

    if password == "" && apiKey == "" {
        return "", nilOptions, ErrNoPassword
    }

    ao := gophercloud.AuthOptions{
        Username: username,
        Password: password,
        ApiKey: apiKey,
        TenantId: tenantId,
        TenantName: tenantName,
    }

    return provider, ao, nil
}

func build() (action int, err error) {
    var (
        provider string
        authOptions gophercloud.AuthOptions
        apiCriteria gophercloud.ApiCriteria
        access *gophercloud.Access
        csp gophercloud.CloudServersProvider
    )

    utils.Debugf("Authenticating with the RackspaceCloud.")

    // Create our auth options set by the user's environment
    provider, authOptions, err = getAuthOptions()
    if err != nil {
        return client.RETURN, err
    }

    // Set our API criteria
    apiCriteria, err = gophercloud.PopulateApi("rackspace-us")
    if err != nil {
        return client.RETURN, err
    }

    apiCriteria.Type = "compute"
    apiCriteria.Region = os.Getenv("OS_REGION_NAME")
    if apiCriteria.Region == "" {
        return client.RETURN, fmt.Errorf("No region set. Please set the OS_REGION_NAME environment variable.")
    }

    // Attempt to authenticate
    access, err = gophercloud.Authenticate(provider, authOptions)
    if err != nil {
        return client.RETURN, err
    }

    utils.Debugf("Searching for a suitable host.")

    csp, err = gophercloud.ServersApi(access, apiCriteria)
    if err != nil {
        return client.RETURN, err
    }

    if servers, err := csp.ListServers(); err == nil {
        targets := []gophercloud.Server{}

        for _, server := range servers {
            if strings.HasPrefix(server.Name, "daas-target_") {
                targets = append(targets, server)
            }
        }

        if len(targets) > 0 {
            utils.Debugf("Found hosts for buildout.")
        } else {
            utils.Debugf("No suitable hosts found. Creating one.")
            return createServer(csp)
        }
    } else {
        return client.RETURN, err
    }

    return client.CONTINUE, nil
}

func createServer(csp gophercloud.CloudServersProvider) (action int, err error) {
    newUuid, err := uuid.NewV4()

    if err != nil {
        return client.RETURN, err
    }

    if images, err := csp.ListImages(); err == nil {
        for _, image := range images {
            utils.Debugf("Image found: %s", image)
        }
    }

    if flavors, err := csp.ListFlavors(); err == nil {
        for _, flavor := range flavors {
            utils.Debugf("Flavor found: %s", flavor)
        }
    }

    name := "daas-target_" + newUuid.String()
    utils.Debugf("Generating new host: %s", name)

    newServer, err := csp.CreateServer(gophercloud.NewServer {
        Name: name,
        ImageRef: "aca656d3-dd70-4d7e-a9e5-f12182871cde",
        FlavorRef: "3",
    })

    if err != nil {
        return client.RETURN, err
    }

    utils.Debugf("Waiting for host build to complete: %s", name)

    for {
        if details, err := csp.ServerById(newServer.Id); err == nil {
            if details.Status == "ACTIVE" {
                break
            }

            fmt.Printf(".")
            time.Sleep(1 * time.Second)
        }
    }

    fmt.Printf("\n")
    utils.Debugf("Host build complete for server: %s", newServer.Id)

    return client.CONTINUE, nil
}

func (rp *RaxPlugin) Before(cli *client.DockerCli, cmd string, args []string) (action int, err error) {
    switch cmd {
        case "build":
        case "ps":
            return build()
    }

    return client.CONTINUE, nil
}

func (rp *RaxPlugin) After(cli *client.DockerCli, callError error, cmd string, args []string) (err error) {
    return nil
}

func (tp *RaxPlugin) String() string {
    return tp.name
}
