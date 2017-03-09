PHP NFS Deploy
--

A small application to allow the speedy running of large/modern PHP applications inside docker containers when running on shared storage such as NFS or glusterfs - whilst retaining the ability to keep the files that must be synced (typically for clustering purposes) such as file uploads.

```
NAME:
   deploy - Sets up modern PHP apps to work better when using docker

USAGE:
   deploy [global options] source destination

COMMANDS:
   help, h  Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --config FILE, -c FILE  Load configuration from FILE (default: ".ddply") [$DEPLOY_CONFIG_FILE]
   --debug, -d             Increase verbosity of running messages
   --help, -h              show help
   --version, -v           print the version
```

### Configuration
Configuration of the folders to sync and the ones to link is achieved via the use of a `.ddply` file.

The presence of such a file in the source directory of your application will cause the full contents of the source directory to be copied to the destinaion folder.

Omitting the file will cause a symlink to be created at the destination pointing at the source directory.

Within the `.dpply` file shared directories can be specified that will be linked from their destination locations back to the source locations. The paths as configured are relative to the source directory.

```yaml
shared:
  - app/files
  - app/sessions
```
