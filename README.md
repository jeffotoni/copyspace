# Copyspace 🚀

Copyspace is a powerful and efficient CLI tool for concurrently copying files and directories to DigitalOcean Spaces and AWS S3-compatible buckets. Written in Go, it uses goroutines to maximize transfer speeds and supports multiple concurrent workers. Ideal for backups, migrations, and synchronization in multi-cloud environments.

## 🌟 Main features

• Concurrent upload and download of files and directories
• Full support for DigitalOcean Spaces and AWS S3
• Dynamic configuration of bucket, endpoint and permissions (ACL)
• Automatic detection of the file MIME type
• Advanced error management and colored logs for better visualization
• Control of the number of workers for performance optimization
• Recursive download maintaining the structure of local directories
• Use of contexts for cancellation and synchronization of processes

#### Install

You can install Copyspace easily on any Linux or Mac system by running:

```bash
$ sh -c "$(wget https://raw.githubusercontent.com/jeffotoni/copyspace/refs/heads/master/v1/install.sh -O -)"
```
#### Credenciais

You will need to create a hidden file in your home, .dokeys which needs to contain your credentials.

#### 🛠️ How to run
1. Set up the credentials file (.dokeys) in your home directory.
2. Run the desired command as per the examples above.
3. Monitor the colored log to track progress and errors

```bash
{
    "key":"xxxxxxxxxxxx",
    "secret":"xxxxxxxxxx",
    "endpoint":"https://sfo2.digitaloceanspaces.com",
    "region":"us-east-1",
    "bucket":"your-bucket"
}
```

Upload a single file to the bucket
```bash
copyspace -file /path/to/file.txt -bucket bucket-name
```

Upload an entire directory recursively with 100 concurrent workers
```bash
copyspace -file /path/to/directory -bucket bucket-name -worker 100
```

Set public permission for the uploaded files
```bash
copyspace -file /path/to/file.txt -bucket bucket-name -acl public
```

Enable recursive download mode from a bucket to a local directory
```bash
copyspace -cp -bucket bucket-name -out /path/to/destination
```
Show help with all available options
```bash
copyspace -h
```