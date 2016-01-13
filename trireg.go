package main

import (
  "os"
  "time"
  "fmt"
  "strconv"
  "io/ioutil"
  "encoding/json"
  "net/http"
  "net/http/cookiejar"
  "net/url"
  "github.com/codegangsta/cli"
  "github.com/howeyc/gopass"
)

func main() {
  app := cli.NewApp()
  app.Name = "Trifork timeregistration"
  app.Usage = "Report hours from your commandline"
  app.Version = "0.4-dev"
  app.Flags = []cli.Flag{
    cli.StringFlag{ Name: "host", Value: "https://tidsreg.trifork.com", EnvVar: "TRIREG_HOST", },
    cli.StringFlag{ Name: "username", Usage: "Select username", EnvVar: "USER,TRIREG_USERNAME", },
    cli.StringFlag{ Name: "password", Usage: "Select password", EnvVar: "TRIREG_PASSWORD", },
    cli.BoolFlag{ Name: "verbose", Usage: "Be more verbose" },
    cli.BoolFlag{ Name: "dryrun", Usage: "Don't submit anything" },
  }
  app.Commands = []cli.Command{
    {
      Name: "hours",
      Usage: "Register hours",
      Flags: []cli.Flag {
        cli.StringFlag{ Name: "date", Value: time.Now().Format("2006-01-02"), Usage: "Select date, format: yyyy-mm-dd", },
        cli.StringFlag{ Name: "customer", Usage: "Select customer", EnvVar: "TRIREG_CUSTOMER,TRIREG_HOURS_CUSTOMER", },
        cli.StringFlag{ Name: "project", Usage: "Select project", EnvVar: "TRIREG_PROJECT,TRIREG_HOURS_PROJECT", },
        cli.StringFlag{ Name: "phase", Usage: "Select phase", EnvVar: "TRIREG_PHASE,TRIREG_HOURS_PHASE", },
        cli.StringFlag{ Name: "activity", Usage: "Select activity", EnvVar: "TRIREG_ACTIVITY,TRIREG_HOURS_ACTIVITY", },
        cli.StringFlag{ Name: "kind", Usage: "Select kind", EnvVar: "TRIREG_KIND,TRIREG_HOURS_KIND", },
        cli.StringFlag{ Name: "invoice-text", Usage: "Optional: Add invoice text", EnvVar: "TRIREG_INVOICE_TEXT,TRIREG_HOURS_INVOICE_TEXT", },
        cli.StringFlag{ Name: "contact", Usage: "Optional: Add contact name", EnvVar: "TRIREG_CONTACT,TRIREG_HOURS_CONTACT", },
      },
      Action: func(c *cli.Context)  {
        verbose := c.GlobalBool("verbose")
        urlRoot := c.GlobalString("host")
        username := c.GlobalString("username")
        password := c.GlobalString("password")

        if password == "" {
          fmt.Printf("Trireg password:")
          password = string(gopass.GetPasswd())
        }

        if verbose {
          fmt.Printf("Registering hours with arguments:\n")
          fmt.Printf("  host: %s\n", urlRoot)
          fmt.Printf("  username: %s\n", username)
        }

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
        quit := func (exitCode int)  {
          respLogout, err := client.Get(urlRoot + "/api/auth/logout")
          defer respLogout.Body.Close()
          if err != nil {
            fmt.Printf("Failed to logout: %s", err)
            os.Exit(99)
          }
          os.Exit(exitCode)
        }
        panic := func (exitCode int, format string, a ...interface{})  {
          if exitCode > 0 {
            fmt.Printf(format, a...)
            println()
          }
          quit(exitCode)
        }
        if respLogin.StatusCode != 200 {
          panic(1, "Failed to login: Wrong username/password")
        }

        respProjects, err := client.Get(urlRoot + "/api/selector/projects")
        defer respProjects.Body.Close()
        if err != nil { panic(1, "Failed to fetch projects: %s", err) }

        projectsBody, err := ioutil.ReadAll(respProjects.Body)
        if err != nil { panic(1, "Failed read response body: %s", err) }

        var projectsJson interface{}
        err = json.Unmarshal([]byte(projectsBody), &projectsJson)

        var customerId int
        for _,customer := range projectsJson.(map[string]interface{})["Customers"].([]interface{}) {
          if customer.(map[string]interface{})["Name"] == c.String("customer") {
            customerId = int(customer.(map[string]interface{})["Id"].(float64))
          }
        }
        if customerId == 0 { panic(1, "Could not find customer: %s", c.String("customer")) }

        var projectId int
        for _,project := range projectsJson.(map[string]interface{})["Projects"].([]interface{}) {
          if int(project.(map[string]interface{})["ParentId"].(float64)) == customerId && project.(map[string]interface{})["Name"] == c.String("project") {
            projectId = int(project.(map[string]interface{})["Id"].(float64))
          }
        }
        if projectId == 0 {
          panic(1, "Could not find project: %s", c.String("project"))
        }

        respProject, err := client.Get(fmt.Sprint(urlRoot + "/api/selector/projects/", float64(projectId)))
        defer respProject.Body.Close()
        if err != nil { panic(1, "Failed to fetch project: %s", err) }

        projectBody, err := ioutil.ReadAll(respProject.Body)
        if err != nil { panic(1, "Failed to unmarshal json: %s", err) }

        var projectJson interface{}
        err = json.Unmarshal([]byte(projectBody), &projectJson)
        var phaseId int
        for _,phase := range projectJson.(map[string]interface{})["Phases"].([]interface{}) {
          if int(phase.(map[string]interface{})["ParentId"].(float64)) == projectId && phase.(map[string]interface{})["Name"] == c.String("phase") {
            phaseId = int(phase.(map[string]interface{})["Id"].(float64))
          }
        }
        if phaseId == 0 {
          panic(1, "Could not find phase: %s", c.String("phase"))
        }
        var activityId int
        for _,activity := range projectJson.(map[string]interface{})["Activities"].([]interface{}) {
          if int(activity.(map[string]interface{})["ParentId"].(float64)) == phaseId && activity.(map[string]interface{})["Name"] == c.String("activity") {
            activityId = int(activity.(map[string]interface{})["Id"].(float64))
          }
        }
        if activityId == 0 {
          panic(1, "Could not find activity: %s", c.String("activity"))
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
          panic(1, "Could not find kind: %s", c.String("kind"))
        }

        submitHours := func (date string, hours string)  {
          if hours == "" {
            panic(1, "No hours")
          }

          if verbose {
            fmt.Printf("Submitting hours for %s: %s", date, hours)
            println()
          }

          if c.GlobalBool("dryrun") {
            return
          }
          respHours, err := client.PostForm(urlRoot + "/api/hours", url.Values{
            "hours": {hours},
            "activityId": {strconv.Itoa(activityId)},
            "date": {c.String("date")},
            "kindId": {strconv.Itoa(kinds[c.String("kind")])},
            "note": {c.String("invoice-text")},
            "contactName": {c.String("contact")},
          })
          defer respHours.Body.Close()
          if err != nil { panic(1, "Failed to submit hours: %s", err) }
          if respHours.StatusCode != 204 {
            panic(1, "Failed to submit hours")
          }

        }

        lastDate, dateParseErr := time.Parse("2006-01-02", c.String("date"))
        if dateParseErr != nil {
          panic(1, "Failed to parse date '%s'", c.String("date"))
        }
        for i := 0; i < len(c.Args()); i++ {
          deltaI := len(c.Args()) - i - 1
          date := lastDate.Add(time.Hour * time.Duration(24 * deltaI * -1))
          submitHours(date.Format("2006-01-02"), c.Args().Get(i))
        }

        quit(0)
      },
    },
  }
  err := app.Run(os.Args)
  if err != nil {
    fmt.Println(err)
  }
}
