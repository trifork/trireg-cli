Commandline interface for Trifork Trireg
===

Usage
---

```
$ trireg help
NAME:
   Trifork timeregistration - Report hours from your commandline

USAGE:
   trireg [global options] command [command options] [arguments...]
   
VERSION:
   0.0.0
   
COMMANDS:
   hours	Register hours
   help, h	Shows a list of commands or help for one command
   
GLOBAL OPTIONS:
   --host "https://tidsreg.trifork.com"	 [$TRIREG_HOST]
   --username 				Select username [$USER, $TRIREG_USERNAME]
   --password 				Select password [$TRIREG_PASSWORD]
   --help, -h				show help
   --version, -v			print the version
   
```

```
$ trireg help hours
NAME:
   trireg hours - Register hours

USAGE:
   trireg hours [command options] [arguments...]

OPTIONS:
   --date "2016-01-11"	Select date, format: yyyy-mm-dd
   --customer 		Select customer [$TRIREG_CUSTOMER, $TRIREG_HOURS_CUSTOMER]
   --project 		Select project [$TRIREG_PROJECT, $TRIREG_HOURS_PROJECT]
   --phase 		Select phase [$TRIREG_PHASE, $TRIREG_HOURS_PHASE]
   --activity 		Select activity [$TRIREG_ACTIVITY, $TRIREG_HOURS_ACTIVITY]
   --kind 		Select kind [$TRIREG_KIND, $TRIREG_HOURS_KIND]
   --invoice-text 	Optional: Add invoice text [$TRIREG_INVOICE_TEXT, $TRIREG_HOURS_INVOICE_TEXT]
   --contact 		Optional: Add contact name [$TRIREG_CONTACT, $TRIREG_HOURS_CONTACT]
```

For instance to register 8 hours on Internal time run the following

```
$ trireg hours \
  --customer="Trifork Ltd."\
  --project="Internal time"\
  --phase="Internal time"\
  --activity="Internal time other"\
  --kind="Not billable"\
  8
```

Rather conveniently trireg can be configured with Environment variables

```
export TRIREG_USERNAME=$USER
export TRIREG_PASSWORD=$(cat ~/.trifork/password)
export TRIREG_CUSTOMER="Trifork Ltd."
export TRIREG_PROJECT="Internal time"
export TRIREG_PHASE="Internal time"
export TRIREG_ACTIVITY="Internal time other"
export TRIREG_KIND="Not billable"
```

With that configured you can run the following command on a daily basis

```
trireg hours 8
```
