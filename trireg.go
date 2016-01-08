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
  app.Commands = []cli.Command{
    {
      Name: "hours",
      Usage: "Register hours",
      Flags: []cli.Flag {
        cli.StringFlag{
          Name: "username",
          Usage: "Select username",
          EnvVar: "TRIREG_USERNAME",
        },
        cli.StringFlag{
          Name: "password",
          Usage: "Select password",
          EnvVar: "TRIREG_PASSWORD",
        },
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
        jar, err := cookiejar.New(nil)
        if err != nil {
          fmt.Printf("Failed to create cookie jar: %s", err)
          os.Exit(1)
        }
        client := http.Client{Jar: jar}

        username := c.String("username")
        password := c.String("password")
        respLogin, err := client.PostForm("https://tidsreg.trifork.com/api/auth/login", url.Values{"username": {username}, "password": { password }})
        defer respLogin.Body.Close()
        if err != nil {
          fmt.Printf("Failed to login: %s", err)
          os.Exit(1)
        }
        //TODO: Check if login is successful

        respProjects, err := client.Get("https://tidsreg.trifork.com/api/selector/projects")
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

        fmt.Println("CustomerId:", customerId)

        var projectId int
        for _,project := range projectsJson.(map[string]interface{})["Projects"].([]interface{}) {
          if int(project.(map[string]interface{})["ParentId"].(float64)) == customerId && project.(map[string]interface{})["Name"] == c.String("project") {
            projectId = int(project.(map[string]interface{})["Id"].(float64))
          }
        }
        fmt.Println("ProjectId:", projectId)
        //TODO: exit if project is null

        respProject, err := client.Get(fmt.Sprint("https://tidsreg.trifork.com/api/selector/projects/", float64(projectId)))
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
        fmt.Println("PhaseId:", phaseId)
        var activityId int
        for _,activity := range projectJson.(map[string]interface{})["Activities"].([]interface{}) {
          if int(activity.(map[string]interface{})["ParentId"].(float64)) == phaseId && activity.(map[string]interface{})["Name"] == c.String("activity") {
            activityId = int(activity.(map[string]interface{})["Id"].(float64))
          }
        }
        fmt.Println("ActivityId:", activityId)

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
        println("Kind:", kinds[c.String("kind")])
        hours := c.Args().First()

        respHours, err := client.PostForm("https://tidsreg.trifork.com/api/hours", url.Values{
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

        println("Add", hours, "hours to project", c.String("customer"), c.String("project"), c.String("phase"), c.String("activity"), c.String("kind"))
      },
    },
  }
  app.Run(os.Args)
}
