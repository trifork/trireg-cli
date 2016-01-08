package main

import (
  "os"
  "fmt"
  "strconv"
  "io/ioutil"
  "encoding/json"
  "net/http"
  "net/http/cookiejar"
  "net/url"
  "github.com/codegangsta/cli"
)

func main() {
  app := cli.NewApp()
  app.Name = "Trifork timeregistration"
  app.Usage = "Report hours from your commandline"
  app.Flags = []cli.Flag{
    cli.StringFlag{
      Name: "host",
      Value: "https://tidsreg.trifork.com",
      EnvVar: "TRIREG_HOST",
    },
    cli.StringFlag{
      Name: "username",
      Usage: "Select username",
      EnvVar: "USER,TRIREG_USERNAME",
    },
    cli.StringFlag{
      Name: "password",
      Usage: "Select password",
      EnvVar: "TRIREG_PASSWORD",
    },
  }
  app.Commands = []cli.Command{
    {
      Name: "hours",
      Usage: "Register hours",
      Flags: []cli.Flag {
        cli.StringFlag{
          Name: "date",
          Value: "2016-01-07",
          Usage: "Select date, format: yyyy-mm-dd",
        },
        cli.StringFlag{
          Name: "customer",
          Usage: "Select customer",
          EnvVar: "TRIREG_CUSTOMER,TRIREG_HOURS_CUSTOMER",
        },
        cli.StringFlag{
          Name: "project",
          Usage: "Select project",
          EnvVar: "TRIREG_PROJECT,TRIREG_HOURS_PROJECT",
        },
        cli.StringFlag{
          Name: "phase",
          Usage: "Select phase",
          EnvVar: "TRIREG_PHASE,TRIREG_HOURS_PHASE",
        },
        cli.StringFlag{
          Name: "activity",
          Usage: "Select activity",
          EnvVar: "TRIREG_ACTIVITY,TRIREG_HOURS_ACTIVITY",
        },
        cli.StringFlag{
          Name: "kind",
          Usage: "Select kind",
          EnvVar: "TRIREG_KIND,TRIREG_HOURS_KIND",
        },
      },
      Action: func(c *cli.Context)  {
        urlRoot := c.GlobalString("host")
        username := c.GlobalString("username")
        password := c.GlobalString("password")

        jar, err := cookiejar.New(nil)
        if err != nil {
          fmt.Printf("Failed to create cookie jar: %s", err)
          os.Exit(1)
        }
        client := http.Client{Jar: jar}

        respLogin, err := client.PostForm(urlRoot + "/api/auth/login", url.Values{"username": {username}, "password": { password }})
        defer respLogin.Body.Close()
        if err != nil {
          fmt.Printf("Failed to login: %s", err)
          os.Exit(1)
        }
        if respLogin.StatusCode != 200 {
          println("Failed to login: Wrong username/password")
          os.Exit(1)
        }

        respProjects, err := client.Get(urlRoot + "/api/selector/projects")
        defer respProjects.Body.Close()
        if err != nil {
          fmt.Printf("Failed to fetch projects: %s", err)
          os.Exit(1)
        }
        projectsBody, err := ioutil.ReadAll(respProjects.Body)
        if err != nil {
          fmt.Printf("Failed to login: %s", err)
          os.Exit(1)
        }
        var projectsJson interface{}
        err = json.Unmarshal([]byte(projectsBody), &projectsJson)

        var customerId int
        for _,customer := range projectsJson.(map[string]interface{})["Customers"].([]interface{}) {
          if customer.(map[string]interface{})["Name"] == c.String("customer") {
            customerId = int(customer.(map[string]interface{})["Id"].(float64))
          }
        }
        if customerId == 0 {
          println("Could not find customer:", c.String("customer"))
          os.Exit(1)
        }

        var projectId int
        for _,project := range projectsJson.(map[string]interface{})["Projects"].([]interface{}) {
          if int(project.(map[string]interface{})["ParentId"].(float64)) == customerId && project.(map[string]interface{})["Name"] == c.String("project") {
            projectId = int(project.(map[string]interface{})["Id"].(float64))
          }
        }
        if projectId == 0 {
          println("Could not find project:", c.String("project"))
          os.Exit(1)
        }

        respProject, err := client.Get(fmt.Sprint(urlRoot + "/api/selector/projects/", float64(projectId)))
        defer respProject.Body.Close()
        if err != nil {
          fmt.Printf("Failed to fetch project: %s", err)
          os.Exit(1)
        }
        projectBody, err := ioutil.ReadAll(respProject.Body)
        if err != nil {
          fmt.Printf("Failed to unmarshal json: %s", err)
          os.Exit(1)
        }
        var projectJson interface{}
        err = json.Unmarshal([]byte(projectBody), &projectJson)
        var phaseId int
        for _,phase := range projectJson.(map[string]interface{})["Phases"].([]interface{}) {
          if int(phase.(map[string]interface{})["ParentId"].(float64)) == projectId && phase.(map[string]interface{})["Name"] == c.String("phase") {
            phaseId = int(phase.(map[string]interface{})["Id"].(float64))
          }
        }
        if phaseId == 0 {
          println("Could not find phase:", c.String("phase"))
          os.Exit(1)
        }
        var activityId int
        for _,activity := range projectJson.(map[string]interface{})["Activities"].([]interface{}) {
          if int(activity.(map[string]interface{})["ParentId"].(float64)) == phaseId && activity.(map[string]interface{})["Name"] == c.String("activity") {
            activityId = int(activity.(map[string]interface{})["Id"].(float64))
          }
        }
        if activityId == 0 {
          println("Could not find activity:", c.String("activity"))
          os.Exit(1)
        }

        kinds := map[string]int{
          "Billable": 13,
          "Overtime 50 %": 14,
          "Overtime 100 %": 15,
          "Not billable": 16,
          "Kilometer": 17,
          "Kilometer not Billable": 19,
          "Voucher": 20,
          "Transport hours": 21,
        }
        kindId := kinds[c.String("kind")]
        if kindId == 0 {
          println("Could not find kind:", c.String("kind"))
          os.Exit(1)
        }

        hours := c.Args().First()

        if hours == "" {
          println("No hours")
          os.Exit(1)
        }

        respHours, err := client.PostForm(urlRoot + "/api/hours", url.Values{
          "hours": {hours},
          "activityId": {strconv.Itoa(activityId)},
          "date": {c.String("date")},
          "kindId": {strconv.Itoa(kinds[c.String("kind")])},
        })
        defer respHours.Body.Close()
        if err != nil {
          fmt.Printf("Failed to submit hours: %s", err)
          os.Exit(1)
        }
        if respHours.StatusCode != 204 {
          println("Failed to submit hours")
          os.Exit(1)
        }
      },
    },
  }
  app.Run(os.Args)
}
