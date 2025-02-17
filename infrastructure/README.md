# image2ascii Infrastructure

## Amazon Web Services (AWS)

This application is hosted on AWS, utilizing many AWS services, including:

- **Amazon Elastic Compute Cloud (Amazon EC2):** The production application is currently running within an AWS EC2 instance. Basic instance details include:
  - **Instance type:** t2 micro
  - **OS:** Ubuntu 24.04.2 LTS
- **Amazon Route 53:** Route 53 was used to purchase the `image2ascii.net` domain, where I established a hosting zone for my application, where DNS records are managed.
- **AWS Identity and Access Management (IAM):** IAM roles were used to establish a secure deployment pipeline between GitHub, AWS, and the CodeDeploy agent.
- **AWS CodeDeploy:** CodeDeploy is used to automate code deployment. When I push a Pull Request to `main`, a GitHub action is activated, which securly triggers the CodeDeploy agent. View `appspec.yml` to see the workflow that's triggered. To view the actual command that is executed to activate the CodeDeploy agent, see the end of `.github/workflows/deploy.yml`.
 
## Files

### image2ascii.service

This is the `systemd` file used to establish the Linux service that runs `app.go` in the background.

**Actual file name:** image2ascii.service

**File location:** /etc/systemd/system

**Commands**:
```bash
# Start the service
systemctl start image2ascii

# Automatically start service on boot
systemctl enable image2ascii
```

### logrorate.conf

This is the `logrotate` configuration file that specifies the log rotate details for my application.

**Actual file name:** image2ascii

**File location:** /etc/logrorate.d

### nginx.conf

This is my `nginx` configuration, which sets up a reverse proxy listening on ports `80` & `443` for internet traffic, and forwards the traffic to my web application running on port `8080`. This configuration file serves many purposes, including:

- Rate limiting for both the website, as well as the ASCII API
- Establishing client & server timeouts
- Added request & response headers for additional security
- Redirecting HTTP requests on port `80` to port `443`
- SSL/TLS encryption managed by Certbot for secure HTTPS connections

**Actual file name:** image2ascii

**File location:** /etc/nginx/sites-available

**Commands:**

```bash
# Create a symlink to enable custom configuration
ln -sf /etc/nginx/sites-available/image2ascii /etc/nginx/sites-enabled/image2ascii

# Restart nginx to enable service
sudo service nginx restart