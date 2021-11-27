# PRISMA Client Installation

These instructions are for installing the PRISMA client on a workstation that already has Windows installed.

## System Requirements

* Windows 10
* Access to the internet for the map tiles.
* Access to the PRISMA RCC PRIMARY and SECONDARY application servers where `twebd` and `tauthd` are running.

## Prepping the System

As of v1.7.6, you will need to prep the system for every user account that will be using the PRISMA Client.

### Determine User Accounts and Create .c2 Directories

Currently, the way the client configures itself to reach the server, you will need to determine exactly which user accounts will use the PRISMA client. For each account you will need to create a `.c2` directory in each users home directory.

The home directories are usually for Windows in `C:\\Users\<username>`. Since Windows Explorer does not like to create dot directories that are prefixed with a period, you will need to open the Windows PowerShell to create the directories.

If you are logged in as the user you want to create the `.c2` directory for, you can just type the following and hit enter in PowerShell:

```
mkdir ~\.c2
```

For other accounts, you will need to use the full path:

```
mkdir C:\\Users\<username>\.c2
```

Once you have created the `.c2` directories you will be able to open them using Windows Explorer.

!!! note
    ~\ and C:\\Users\<username> are the same. The ~\ is just a shortcut to the system so you don't have to write C:\\Users\<username> every time. If you need to reference a user home directory other than the user you are logged in as, you will have to use the full path as ~ only works or the current logged in user.

## Configure the Client

PRISMA RCC has a custom configuration file, which the system relies on to find the PRISMA RCC server that the client will be connecting to. This file will be located in the `.c2` directory FOR EVERY USER that will be running the PRISMA client.

The file is named `production.json` so you will create `production.json` in every users .c2 folder:

```
C:\\Users\<username>\.c2\production.json
```

!!! note Helpful Tips
    When creating this configuration file, we recommend creating the file for a single user, logging in as that user then installing the client in the following section to verify the config is correct. Once the configuration is confirmed to work, then copy it to all client Windows systems and all users. The PRISMA team keeps copies of all deployed configuration in the prisma-c2 bucket under `client-configurations` so be sure to check with the team before creating a new config. Ideally, the installation teams should also save copies of all configs as well for every deployment for backup and ease of setting up new systems.

!!! note For Developers
    When using the client on a development system, or running code build locally, you will need to use a config file called `development.json` instead of `production.json`. The app picks which file to load based on the current node environment (NODE_ENV) being set to `production` or `development`.

Now, create or copy and existing config file into the `.c2` folder. Open the file using Notepad++ or Notepad and verify the IP address match the IPs for the Primary and Secondary Application Server.
Typical client configuration file looks like the following:

```json tab="With Replication Server"
{
    "configurations": [{
        "default": true,
        "name": "PRISMA RCC Primary",
        "url": "http://<IPADDR PRIMARY>:8081/api/v2/config.json"
    },{
        "default": false,
        "name": "PRISMA RCC Secondary",
        "url": "http://<IPADDR SECONDARY>:8081/api/v2/config.json"
    }]
}
```


```json tab="Single Server"
{
    "configurations": [{
        "default": true,
        "name": "PRISMA RCC",
        "url": "http://<IPADDR>:8081/api/v2/config.json"
    }]
}
```

The config file URL path is is always `:8081/api/v2/config.json`.

!!! note Switch Servers
    If you need to switch which server the client is using, then swap the default: true lines to turn the PRIMARY to false and the SECONDARY to true. Then restart the application. Never have more than 1 listed as default: true.

## PRISMA Client Installation

Now that the configuration is setup, you can install the client.

Download the PRISMAInstaller-X.X.X.exe client installer. Double click the installer to install the client. Once complete, the installer will automatically start the application.

You should now have a Start Menu link and Desktop link to PRISMA, under the `Orolia` group name.

!!! warning
    you may get an error that the configuration could not be found. This means the install suceeded,
    and Windows tried to start the client that was installed but it failed to that. This is ok, once you
    have created the client `production.json` the client should load correctly.

!!! warning
    If the application launches but the login screen says it can't connect one of three issues may have happened. The application on the server isn't running, the production.json is invalid or has the wrong server URL, or the SSL certificate has not be installed (see [TLS Certificates](#TLSCertificates) below for instructions.

At this point, it is convention to also pin the application to the task bar at the bottom. When the application is running, right click on the PRISMA icon on the bottom task bar (next to the start menu) and select PIN. This allows the PRISMA icon to always be located on the task bar, even if the application isn't running for easier launching.

### Add Secondary Server Shortcuts

To make it simpler on the user during a failover scenario, we will also install a small script and a desktop shortcut to start a client connected to the secondary server. This setup allows the user to just close the app and click on the secondary icon to switch to the secondary server during a failover instead of the procedure of opening the production.json and changing the default property value.

First, create the following file (or copy an existing file from our `prisma-c2/client configurations/startSecondaryClient.bat`) as `C:\\Users\<username>\.c2\startSecondaryClient.bat`. The bat file should contain the following (be sure to change the `<USERNAME>` to the correct IP of the secondary server and the `<USERNAME>` to the correct username of the user:

```
start /wait C:\Users\<USERNAME>\AppData\Local\PRISMA\PRISMA.exe -- --config httpe://<IPADDRESS>:8081/api/v2/config.json
exit
```

Now, right click the file and select `Create Shortcut`. Rename the new shortcut `PRISMA Secondary` and then copy the new shortcut to the Desktop and place it next to the current PRISMA shortcut. Then, right click again and edit the info to add an icon image. Set the icon image to the icon from [https://releases.mcmurdo.io/images/logo/app_icon.ico](https://releases.mcmurdo.io/images/logo/app_icon.ico).

!!! note
    You must do this procedure for every user who is going to start the client.

## TLS Certificates

When using a self signed certificate the workstation needs be be configured to trust it. Download the certificate by visiting the following URL In a web browser: [`https://<server ipAddress>:8080/api/v2/certificate.pem`](https://localhost:8080/api/v2/certificate.pem)
where ipAddress is the IP address of the application serrver. The web browser should show a warning that the site is not secure. Continue and save the certificate to the Desktop.

#### Windows SSL Certificate Install

For Windows users, from the Start Menu, search for and run the "mmc" application. From the menu, select File => Add/Remove Snap In. Move "Certificates" from the available snap-ins to the selected snap-ins. When prompted on how this snap-in will manage certificates, select "Computer account". Then select "Local computer" for the computer the snap-in will manage. Click "OK" to add the snap-in.

From the tree in the left frame, select on Console Root => Trusted Root Certification Authorities => Certificates. Right click on Certificates and then Select All Tasks => Import and click on Next. Click on "Browse". In the lower right hand corner of the open file dialog, change the drop-down from "X.509 Certificate" to "All Files". Select "certificate.pem". Click Next when asked about the Certificate Store and then click the "Finish" button. Verify that the certificate named "Orolia PRISMA PRISMA" has been added to the list of certificates.

#### For macOS Users

For Mac Users, double click the pem file that was downloaded. It should open Keychain application and ask for a password. Find the certificate under Orolia Prisma C2 (`<ip or localhost>`) and right click on it then select Get Info. In the window that opens expand the Trust panel then under When using this certificate, select Always Trust then close the window. This should ask for a password.


#### Verify SSL Installation

Restart the web browser and verify the certificate is now trusted by visiting the following URL: [`http://<server ipAddress>:8080/api/v2/apidocs.json`](http://localhost:8080/api/v2/apidocs.json)

The browser should now indicate that the connection is secure and you should see something like the following:

```
"version": "1.8.1"
```

## Initial Login
Start PRISMA by clicking on the Desktop icon. Default credentials are:
 * Username: admin
 * Password: admin

Change the adminsitrator password at this time. Click users icon on the navigation bar on the left,
then click `admin` user and change the password.

## Cleanup

Remove any installation artifacts remaining on the Desktop.


## SARMAP

If the installation requires SARMAP, then these instructions will get the PRISMA installation talking to the SARMAP install.

### Prerequisites

SARMAP is installed with the correct version of SARMAP that works with PRISMA. SARMAP has the details on what version this is. We do have some copies of the installs in the aws bucket `prisma-c2`.

### Configuring

To get the SARMAP install to read the beacons added to an incident in PRISMA, you need to copy or edit the `LiveLayers.xml` file.

See [SARMAP Install Guide for procedures.](../tutorials/sarmap-c2.md)

## Failover

The following section describes how the frontend needs to be reconfigured when failover occurs and the secondary server assumes responsibility for the application.

The easiest way to start using the failover server is to close the application and use the PRISMA Secondary desktop icon to load a client connected to the secondary server.

If this icon is not available, then you will have to open the `C:\\Users\<USERNAME>\.c2\production.json` and change the the `default: true` line to `defdault: false` then
ensure the secondary server ip has the `default: true` line. Then restart the PRISMA Application.
