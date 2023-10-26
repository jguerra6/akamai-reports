# Akamai API reporting
## How to use the cli
1. Clone the project
2. Create your credentials
3. Add the credentials to your config file
4. Create the CPCode Map
5. Run the report

### Credentials

Create API credentials to manage all the accounts by using [Akamai IAM](https://techdocs.akamai.com/developer/docs/manage-many-accounts-with-one-api-client).

### Create your config file
From the previous step, download the credentials and save them under the config folder as `.edgerc`

Your file should look like this:
```
[default]
client_secret = your_info_here
host = your_info_here
access_token = your_info_here
client_token = your_info_here
```

### Create the CPCode Map
Run the program with the `init` command:

```
go run main.go init
```

This will create a folder `data` with two json files, containing the full list of Account Switch keys and another with the full list of the CPCodes per account.

### Run the report

Currently, it's only supported that the report will run against all the accounts except **Brightcove Inc._Value Added Reseller** since it contains all the CPCodes for house traffic and several test ones it causes the request to take longer and fail.

Run the program with the `report` command and required parameters:

```
go run main.go report -start "2023-09-01" -end "2023-09-27"
```

The required params are:

| param |                     Description                      |
|-------|:----------------------------------------------------:|
| start | The start date for the report in `YYYY-MM-DD format` |
| end   |  The end date for the report in `YYYY-MM-DD format`  |

This will produce a new folder called `report` inside the data folder which will contain the `report.csv` file with the information of all the CPCodes.
If all the data of a day is 0 it will be excluded so if you don't see a day for a specific CPCode that could be the reason.

### What's next?
* Improve error logging
* Add custom separators as a flag

