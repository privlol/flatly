# flatly

Flatly dynamically checks a static config file installing and or removing packages.

## Installation

1. Clone the repository `https://github.com/privlol/flatly.git`
2. Enter the repo `cd Flatly`
3. Build the project `make build`
4. Install `make install`
6. Enable the service or run the daemon in CLI `systemctl --user enable flatly.service` or `flatly daemon`

## Usage

### Quick commands
These command's are just for simple interaction with Flatly

#### Install Package

To install a Flatpak package, use the following command:
`flatly add <package_name>`
Replace <package_name> with the name of the Flatpak package you want to install.

#### Remove Package

To remove a Flatpak package, use:
`flatly remove <package_name>`
Replace <package_name> with the name of the Flatpak package you want to remove.

#### Backup

Flatly automatically creates backups of the active.json file before making any changes. You can also manually back it up by running:
`flatly backup`
This will create a backup of the current list of installed Flatpak packages.
