# MailBrief Installation Guide

This guide walks you through deploying MailBrief from scratch. It assumes you are starting with nothing more than a computer and an internet connection. Every step is explained, every command is provided, and placeholder values are clearly marked so you know what to replace with your own information.

**What is MailBrief?** It is an AI-powered email digest dashboard. You sign up, get a managed inbox (like `yourname@yourdomain.com`), forward your newsletters to it, and receive a daily AI-generated summary of everything you received.

**How long will this take?** Expect 2 to 3 hours for a first-time setup. Most of that time is waiting for DNS records to propagate and verifying each piece works.

---

## Table of Contents

1. [Part 1: Prerequisites](#part-1-prerequisites)
2. [Part 2: Domain Setup (Cloudflare)](#part-2-domain-setup-cloudflare)
3. [Part 3: Database Setup (Neon.tech)](#part-3-database-setup-neontech)
4. [Part 4: Backend Deployment (Google Cloud Run)](#part-4-backend-deployment-google-cloud-run)
5. [Part 5: Frontend Deployment (Vercel)](#part-5-frontend-deployment-vercel)
6. [Part 6: Domain Configuration (Cloudflare to Vercel)](#part-6-domain-configuration-cloudflare-to-vercel)
7. [Part 7: Email Sending Setup (Resend)](#part-7-email-sending-setup-resend)
8. [Part 8: Email Routing (Cloudflare Email Workers)](#part-8-email-routing-cloudflare-email-workers)
9. [Part 9: Digest Scheduler (Google Cloud Scheduler)](#part-9-digest-scheduler-google-cloud-scheduler)
10. [Part 10: Verification and Testing](#part-10-verification-and-testing)
11. [Part 11: Ongoing Maintenance](#part-11-ongoing-maintenance)
12. [Part 12: Troubleshooting](#part-12-troubleshooting)

---

## Part 1: Prerequisites

Before you begin, you need to create several accounts and install a few tools on your computer. None of these cost money to start -- they all have free tiers sufficient for a personal deployment.

### Accounts to Create

Create an account on each of these services. You only need a free account for each one.

| Service | URL | What it does |
|---------|-----|--------------|
| **GitHub** | [github.com](https://github.com) | Hosts the source code |
| **Cloudflare** | [cloudflare.com](https://cloudflare.com) | Manages your domain and routes email |
| **Neon.tech** | [neon.tech](https://neon.tech) | Hosts the PostgreSQL database |
| **Google Cloud** | [console.cloud.google.com](https://console.cloud.google.com) | Runs the backend service and cron scheduler |
| **Vercel** | [vercel.com](https://vercel.com) | Hosts the frontend website |
| **Resend** | [resend.com](https://resend.com) | Sends digest emails to users |
| **Anthropic** | [console.anthropic.com](https://console.anthropic.com) | Provides the AI that summarizes emails |

> **Note:** Google Cloud requires a billing account with a credit card, but the free tier is generous and a small personal deployment will likely cost nothing or very little.

### Tools to Install

You need the following tools installed on your computer. Open your terminal (on macOS: search for "Terminal" in Spotlight; on Windows: use PowerShell or WSL) and follow the instructions below.

#### 1. Git

Git is a version control tool that lets you download the source code.

**macOS:**
```bash
xcode-select --install
```
A dialog box will appear. Click "Install" and wait for it to finish.

**Windows (WSL):**
```bash
sudo apt update && sudo apt install git
```

**Verify it works:**
```bash
git --version
```
You should see something like `git version 2.39.0`.

#### 2. Node.js 22+

Node.js runs the frontend application. You need version 22 or newer.

**macOS (using Homebrew):**
```bash
# Install Homebrew first if you do not have it:
/bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)"

# Then install Node.js:
brew install node@22
```

**Windows (WSL):**
```bash
curl -fsSL https://deb.nodesource.com/setup_22.x | sudo -E bash -
sudo apt install -y nodejs
```

**Verify it works:**
```bash
node --version
```
You should see `v22.x.x` or higher.

#### 3. Go 1.24+

Go is the programming language the backend is written in.

**macOS:**
```bash
brew install go
```

**Windows (WSL) / Linux:**

Download from [go.dev/dl](https://go.dev/dl/) and follow the instructions for your operating system.

**Verify it works:**
```bash
go version
```
You should see `go1.24.x` or higher.

#### 4. Docker

Docker is used to build the backend into a container for deployment.

Download and install Docker Desktop from [docker.com/products/docker-desktop](https://www.docker.com/products/docker-desktop/). Open Docker Desktop after installation and wait for it to start (you will see a green "running" indicator).

**Verify it works:**
```bash
docker --version
```

#### 5. Google Cloud CLI (gcloud)

The `gcloud` command line tool lets you deploy to Google Cloud from your terminal.

**macOS:**
```bash
brew install google-cloud-sdk
```

**Windows (WSL) / Linux:**
```bash
curl https://sdk.cloud.google.com | bash
```
Restart your terminal after installation.

**Initialize gcloud:**
```bash
gcloud init
```
This will open your browser and ask you to log in with your Google account. Follow the prompts to:
1. Log in with the Google account you used for Google Cloud.
2. Select your Google Cloud project (or create one -- name it something like `mailbrief`).
3. Choose a default region (e.g., `us-central1`).

**Verify it works:**
```bash
gcloud --version
```

---

## Part 2: Domain Setup (Cloudflare)

You need a domain name (like `mailbrief.io` or `mydigest.com`). This domain will be used for your website and for email addresses.

### Buy a Domain on Cloudflare Registrar

1. Log in to [dash.cloudflare.com](https://dash.cloudflare.com).
2. In the left sidebar, click **Domain Registration**, then **Register Domains**.
3. Type the domain name you want in the search box (e.g., `mydigest.com`) and press Enter.
4. Cloudflare will show you available domain names and their prices. Most `.com` domains cost around $10/year.
5. Click **Purchase** next to the domain you want.
6. Fill in your contact information and complete the payment.
7. After purchase, your domain will appear in your Cloudflare dashboard. Click on it to see its settings.

> **Why Cloudflare?** You will use Cloudflare for three things: domain registration, DNS management, and email routing. Having all three in one place makes the setup simpler.

Write down your domain name. This guide will use `yourdomain.com` as a placeholder -- replace it with your actual domain everywhere you see it.

---

## Part 3: Database Setup (Neon.tech)

MailBrief stores everything -- user accounts, emails, digests, settings -- in a PostgreSQL database. Neon.tech provides a free, managed PostgreSQL database in the cloud.

### Create a Neon Project

1. Go to [console.neon.tech](https://console.neon.tech) and sign in.
2. Click **New Project**.
3. Give your project a name: `mailbrief`.
4. Choose a region close to where your users will be (e.g., `US East` for North America).
5. Click **Create Project**.

### Get Your Connection String

1. After creation, Neon will show you a connection string. It looks like this:
   ```
   postgresql://neondb_owner:abc123xyz@ep-cool-name-123456.us-east-2.aws.neon.tech/neondb?sslmode=require
   ```
2. **Copy this entire string and save it somewhere safe.** You will need it in both Part 4 (backend) and Part 5 (frontend).
3. If you need to find it later: go to your Neon project dashboard, click **Connection Details** in the sidebar, and copy the connection string.

> **Important:** This connection string contains your database password. Treat it like a password -- do not share it publicly or commit it to source code.

> **Why one database?** Both the frontend (user authentication via Better Auth) and the backend (emails, digests, settings) share the same database. This keeps the architecture simple and avoids synchronization issues.

---

## Part 4: Backend Deployment (Google Cloud Run)

The backend is a Go service that receives emails, processes them with AI, and generates digests. You will deploy it to Google Cloud Run, which runs your code in a container and scales automatically.

### Step 1: Enable Required Google Cloud APIs

Run these commands to enable the services you need:

```bash
gcloud services enable run.googleapis.com
gcloud services enable cloudbuild.googleapis.com
gcloud services enable artifactregistry.googleapis.com
gcloud services enable cloudscheduler.googleapis.com
```

These commands tell Google Cloud to activate:
- **Cloud Run**: runs your backend container
- **Cloud Build**: builds your Docker container in the cloud
- **Artifact Registry**: stores your built container images
- **Cloud Scheduler**: triggers digest generation on a schedule (used in Part 9)

### Step 2: Clone the Repository

```bash
git clone https://github.com/hadfielj/taran.git
cd taran
```

This downloads all the source code to your computer.

### Step 3: Generate Secrets

You need three random secret strings. Run each of these commands and save the output:

```bash
# Webhook secret -- authenticates email deliveries from Cloudflare
openssl rand -hex 32

# API key -- authenticates communication between frontend and backend
openssl rand -hex 32

# Unsubscribe secret -- secures email unsubscribe links
openssl rand -hex 32
```

Each command outputs a long string of random characters like `a3f8b2c1d4e5...`. Save all three -- you will need them in the next step.

### Step 4: Deploy to Cloud Run

Run the following command, replacing every `<placeholder>` with your actual values:

```bash
cd backend

gcloud run deploy taran \
  --source . \
  --region us-central1 \
  --allow-unauthenticated \
  --port 8080 \
  --set-env-vars "\
TARAN_HOST=0.0.0.0,\
TARAN_PORT=8080,\
TARAN_DB_URL=<your-neon-connection-string>,\
TARAN_WEBHOOK_SECRET=<your-webhook-secret>,\
TARAN_API_KEY=<your-api-key>,\
TARAN_LLM_PROVIDER=anthropic,\
TARAN_ANTHROPIC_API_KEY=<your-anthropic-api-key>,\
TARAN_ANTHROPIC_MODEL=claude-haiku-4-5-20251001,\
TARAN_EMAIL_DOMAIN=<yourdomain.com>,\
TARAN_ADMIN_EMAILS=<your-personal-email>,\
TARAN_ALLOWED_ORIGINS=https://<yourdomain.com>,\
TARAN_UNSUBSCRIBE_SECRET=<your-unsubscribe-secret>,\
TARAN_DEFAULT_MONTHLY_TOKEN_LIMIT=500000"
```

**What each environment variable means:**

| Variable | Description | Example |
|----------|-------------|---------|
| `TARAN_HOST` | IP address to listen on. Always `0.0.0.0` for Cloud Run. | `0.0.0.0` |
| `TARAN_PORT` | Port number. Cloud Run expects `8080`. | `8080` |
| `TARAN_DB_URL` | Your Neon PostgreSQL connection string from Part 3. | `postgresql://neondb_owner:...` |
| `TARAN_WEBHOOK_SECRET` | Random secret that authenticates incoming email webhooks. | Output of `openssl rand -hex 32` |
| `TARAN_API_KEY` | Random secret that authenticates the frontend-to-backend connection. | Output of `openssl rand -hex 32` |
| `TARAN_LLM_PROVIDER` | Which AI provider to use. Set to `anthropic`. | `anthropic` |
| `TARAN_ANTHROPIC_API_KEY` | Your Anthropic API key, from [console.anthropic.com](https://console.anthropic.com) under API Keys. | `sk-ant-api03-...` |
| `TARAN_ANTHROPIC_MODEL` | Which Claude model to use. Haiku is fast and affordable. | `claude-haiku-4-5-20251001` |
| `TARAN_EMAIL_DOMAIN` | Your domain name, used to create user email addresses. | `yourdomain.com` |
| `TARAN_ADMIN_EMAILS` | Comma-separated list of admin email addresses. | `you@gmail.com` |
| `TARAN_ALLOWED_ORIGINS` | Your frontend URL, for CORS security. | `https://yourdomain.com` |
| `TARAN_UNSUBSCRIBE_SECRET` | Random secret that secures email unsubscribe links. | Output of `openssl rand -hex 32` |
| `TARAN_DEFAULT_MONTHLY_TOKEN_LIMIT` | AI token budget per user per month. 500,000 is a good start. | `500000` |

> **Note:** You will add `TARAN_RESEND_API_KEY` and `TARAN_BASE_URL` later in Part 7, after setting up Resend.

The first deploy will take 3 to 5 minutes. Google Cloud will:
1. Upload your source code.
2. Build the Docker container in the cloud.
3. Deploy the container to Cloud Run.

### Step 5: Note Your Cloud Run URL

When the deploy finishes, the output will include a line like:

```
Service URL: https://taran-abc123-uc.a.run.app
```

**Copy this URL and save it.** You will need it in Part 5 (frontend), Part 8 (email routing), and Part 9 (scheduler).

### Step 6: Test the Health Endpoint

```bash
curl https://<your-cloud-run-url>/health
```

Replace `<your-cloud-run-url>` with the URL from the previous step. You should see:

```json
{"status":"healthy"}
```

If you see `{"status":"unhealthy","reason":"database unreachable"}`, double-check your `TARAN_DB_URL` connection string.

> **Getting your Anthropic API key:** Go to [console.anthropic.com](https://console.anthropic.com). Click **API Keys** in the sidebar. Click **Create Key**. Give it a name like "mailbrief" and copy the key. It starts with `sk-ant-`.

---

## Part 5: Frontend Deployment (Vercel)

The frontend is a Next.js application that provides the web interface -- login, inbox, digest viewer, settings, and admin dashboard.

### Step 1: Install the Vercel CLI

```bash
npm install -g vercel
```

### Step 2: Create OAuth Applications

Users log in to MailBrief with their Google or GitHub accounts. You need to register your application with both providers so they allow this.

#### Google OAuth

1. Go to [console.cloud.google.com](https://console.cloud.google.com).
2. Make sure you are in the correct project (the same one you used for `gcloud`).
3. In the search bar at the top, type "OAuth consent screen" and click the result.
4. Click **Get Started** (or **Configure Consent Screen** if you have done this before).
5. Fill in:
   - **App name**: `MailBrief`
   - **User support email**: your email address
   - **Developer contact email**: your email address
6. Click **Save and Continue** through each step. For scopes, you do not need to add any -- the defaults are fine.
7. Now go to **Credentials** in the left sidebar (under APIs and Services).
8. Click **Create Credentials** at the top, then **OAuth 2.0 Client ID**.
9. For **Application type**, select **Web application**.
10. Give it a name: `MailBrief`.
11. Under **Authorized redirect URIs**, click **Add URI** and enter:
    ```
    https://<yourdomain.com>/api/auth/callback/google
    ```
12. Click **Create**.
13. A dialog will show your **Client ID** and **Client Secret**. Copy both and save them.

#### GitHub OAuth

1. Go to [github.com/settings/developers](https://github.com/settings/developers).
2. Click **OAuth Apps**, then **New OAuth App**.
3. Fill in:
   - **Application name**: `MailBrief`
   - **Homepage URL**: `https://<yourdomain.com>`
   - **Authorization callback URL**: `https://<yourdomain.com>/api/auth/callback/github`
4. Click **Register application**.
5. You will see your **Client ID** on the app page.
6. Click **Generate a new client secret**. Copy the secret immediately -- GitHub will only show it once.

### Step 3: Generate the Auth Secret

```bash
openssl rand -hex 32
```

Save the output. This secret is used by Better Auth to encrypt session tokens.

### Step 4: Link the Project to Vercel

```bash
cd frontend
vercel link
```

Vercel will ask you several questions:
- **Set up and develop?** Yes
- **Which scope?** Select your Vercel account
- **Link to existing project?** No (if this is your first time) -- create a new one
- **Project name?** `mailbrief` (or whatever you prefer)
- **Which directory is the code in?** `./` (you are already in the `frontend/` directory)

### Step 5: Add Environment Variables

Go to [vercel.com](https://vercel.com), navigate to your project, then click **Settings** in the top menu, then **Environment Variables** in the sidebar.

Add each of the following variables. For each one, set it for all environments (Production, Preview, Development):

| Variable | Value | Description |
|----------|-------|-------------|
| `BETTER_AUTH_SECRET` | Output of `openssl rand -hex 32` | Encrypts user sessions |
| `BETTER_AUTH_URL` | `https://<yourdomain.com>` | Your site's public URL |
| `DATABASE_URL` | Your Neon connection string from Part 3 | Same database as the backend |
| `GOOGLE_CLIENT_ID` | From Step 2 (Google OAuth) | Google login integration |
| `GOOGLE_CLIENT_SECRET` | From Step 2 (Google OAuth) | Google login integration |
| `GITHUB_CLIENT_ID` | From Step 2 (GitHub OAuth) | GitHub login integration |
| `GITHUB_CLIENT_SECRET` | From Step 2 (GitHub OAuth) | GitHub login integration |
| `BACKEND_URL` | Your Cloud Run URL from Part 4 | Where the frontend sends API requests |
| `API_KEY` | Must match `TARAN_API_KEY` from Part 4 | Authenticates frontend-to-backend requests |
| `NEXT_PUBLIC_APP_NAME` | `MailBrief` | Displayed in the UI |
| `NEXT_PUBLIC_APP_URL` | `https://<yourdomain.com>` | Used for generating links |
| `NEXT_PUBLIC_EMAIL_DOMAIN` | `<yourdomain.com>` | Shown to users as their email domain |
| `ADMIN_EMAILS` | `<your-personal-email>` | Email addresses that get admin access |

> **Important:** `API_KEY` in the frontend must be exactly the same value as `TARAN_API_KEY` in the backend. If they do not match, the frontend will not be able to communicate with the backend.

### Step 6: Deploy

```bash
vercel --prod
```

This builds and deploys the frontend. It takes about 1 to 2 minutes. When it finishes, Vercel will show you a URL like `https://mailbrief.vercel.app`.

> **Alternative: Auto-deploy from GitHub.** Instead of manual deploys, you can connect your GitHub repository to Vercel. Go to [vercel.com/new](https://vercel.com/new), import your GitHub repository, set the root directory to `frontend/`, add the environment variables, and every push to `main` will automatically deploy.

---

## Part 6: Domain Configuration (Cloudflare to Vercel)

Right now your site is accessible at a Vercel-provided URL (like `mailbrief.vercel.app`). This step points your custom domain to Vercel so visitors can reach it at `https://yourdomain.com`.

### Step 1: Add DNS Records in Cloudflare

1. Log in to [dash.cloudflare.com](https://dash.cloudflare.com).
2. Click on your domain.
3. Click **DNS** in the left sidebar, then **Records**.
4. Add two records:

**Record 1 -- points the root domain to Vercel:**
- **Type**: `A`
- **Name**: `@`
- **IPv4 address**: `76.76.21.21`
- **Proxy status**: Click the orange cloud to turn it **off** (grey cloud, DNS only)
- Click **Save**

**Record 2 -- points the www subdomain to Vercel:**
- **Type**: `CNAME`
- **Name**: `www`
- **Target**: `cname.vercel-dns.com`
- **Proxy status**: Click the orange cloud to turn it **off** (grey cloud, DNS only)
- Click **Save**

> **Why disable the Cloudflare proxy (orange cloud)?** Vercel provides its own SSL certificates and CDN. Having Cloudflare's proxy enabled can interfere with Vercel's certificate provisioning.

### Step 2: Add the Domain in Vercel

1. Go to [vercel.com](https://vercel.com) and open your project.
2. Click **Settings** in the top menu.
3. Click **Domains** in the left sidebar.
4. Type your domain (`yourdomain.com`) and click **Add**.
5. Vercel will show instructions. Since you already added the DNS records, it will begin verifying.
6. Also add `www.yourdomain.com` if you want it to redirect to the root domain.

### Step 3: Wait for SSL

Vercel will automatically provision an SSL certificate for your domain. This usually takes 1 to 10 minutes. You can check the status on the Domains settings page -- it will show "Valid Configuration" with a green checkmark when ready.

### Step 4: Verify

Open your browser and visit:
```
https://yourdomain.com
```

You should see the MailBrief login page. If you see a browser security warning about the certificate, wait a few more minutes and try again.

---

## Part 7: Email Sending Setup (Resend)

Resend is the service that sends digest emails to your users. When a digest is generated, the backend uses Resend to deliver it to the user's real email address.

### Step 1: Create a Resend Account

1. Go to [resend.com](https://resend.com) and sign up.

### Step 2: Add Your Domain

1. In the Resend dashboard, click **Domains** in the left sidebar.
2. Click **Add Domain**.
3. Enter your domain: `<yourdomain.com>`.
4. Click **Add**.

### Step 3: Add DNS Records in Cloudflare

Resend will show you a list of DNS records to add. You need to add these to Cloudflare so that Resend is authorized to send email from your domain.

Go back to Cloudflare, click on your domain, then **DNS** > **Records**, and add each record Resend shows you. Typically these are:

**SPF Record (TXT):**
- **Type**: `TXT`
- **Name**: `@`
- **Content**: Resend will provide the exact value (something like `v=spf1 include:send.resend.com ~all`)

> **Note:** If you already have an SPF record (which you might from Part 8), you need to merge them. Do not create two separate SPF TXT records. Instead, combine them into one record. For example: `v=spf1 include:send.resend.com include:_spf.mx.cloudflare.net ~all`

**DKIM Records (CNAME x 3):**

Resend will show you three CNAME records for DKIM. Add each one exactly as shown:
- **Type**: `CNAME`
- **Name**: (Resend provides this, e.g., `resend._domainkey`)
- **Target**: (Resend provides this)

Repeat for all three DKIM records.

**DMARC Record (TXT) -- recommended but optional:**
- **Type**: `TXT`
- **Name**: `_dmarc`
- **Content**: `v=DMARC1; p=none;`

### Step 4: Verify in Resend

Go back to the Resend dashboard. Click on your domain. Click **Verify**. If all DNS records have propagated (this can take up to an hour), the status will change to "Verified".

### Step 5: Get Your API Key

1. In Resend, click **API Keys** in the left sidebar.
2. Click **Create API Key**.
3. Give it a name: `mailbrief`.
4. For permissions, select **Sending access** and restrict it to your domain.
5. Copy the key. It starts with `re_`.

### Step 6: Update the Backend

Now add the Resend API key and the base URL to your Cloud Run deployment:

```bash
gcloud run services update taran \
  --region us-central1 \
  --set-env-vars "\
TARAN_RESEND_API_KEY=<your-resend-api-key>,\
TARAN_BASE_URL=https://<your-cloud-run-url>"
```

> **Warning:** The `--set-env-vars` command on `gcloud run services update` will **replace** all environment variables unless you use `--update-env-vars` instead. To add new variables without removing existing ones, use:
> ```bash
> gcloud run services update taran \
>   --region us-central1 \
>   --update-env-vars "\
> TARAN_RESEND_API_KEY=<your-resend-api-key>,\
> TARAN_BASE_URL=https://<your-cloud-run-url>"
> ```

---

## Part 8: Email Routing (Cloudflare Email Workers)

This is the core of how MailBrief receives email. When someone sends an email to `anything@yourdomain.com`, Cloudflare receives it and forwards it to your backend for processing.

### Step 1: Enable Email Routing

1. In the Cloudflare dashboard, click on your domain.
2. Click **Email** in the left sidebar, then **Email Routing**.
3. Click **Get Started** or **Enable Email Routing**.
4. Cloudflare will ask you to add MX records. These tell email servers to deliver mail for your domain to Cloudflare. Click **Add Records Automatically**, or add them manually:

| Type | Name | Mail server | Priority |
|------|------|-------------|----------|
| MX | `@` | `route1.mx.cloudflare.net` | 69 |
| MX | `@` | `route2.mx.cloudflare.net` | 6 |
| MX | `@` | `route3.mx.cloudflare.net` | 93 |

Cloudflare may also update your SPF TXT record to include `include:_spf.mx.cloudflare.net`. If you already have an SPF record from Part 7, merge them as described earlier.

### Step 2: Create an Email Worker

An Email Worker is a small script that runs every time an email arrives. It will forward the email to your backend.

1. In the Cloudflare dashboard, click **Workers & Pages** in the left sidebar.
2. Click **Create**.
3. Click **Create Worker**.
4. Give it a name: `mailbrief-email-forwarder`.
5. Click **Deploy** (this deploys the default "Hello World" worker -- you will replace the code next).
6. After deployment, click **Edit Code** (or go to the worker and click **Quick Edit**).
7. Delete everything in the editor and paste the following code:

```js
export default {
  async email(message, env) {
    const rawEmail = await new Response(message.raw).text();

    const response = await fetch(env.WEBHOOK_URL, {
      method: "POST",
      headers: {
        "Content-Type": "message/rfc822",
        "X-Webhook-Secret": env.WEBHOOK_SECRET,
        "X-Original-To": message.to,
      },
      body: rawEmail,
    });

    if (!response.ok) {
      message.setReject(`Webhook failed: ${response.status}`);
    }
  }
};
```

8. Click **Save and Deploy**.

**What this code does:** When an email arrives, the worker reads the raw email content and sends it (via an HTTP POST request) to your backend's webhook endpoint. The `X-Webhook-Secret` header authenticates the request so your backend knows it is a legitimate delivery, not spam.

### Step 3: Add Environment Variables to the Worker

1. Go to **Workers & Pages** and click on your `mailbrief-email-forwarder` worker.
2. Click **Settings**, then **Variables and Secrets**.
3. Add two variables:

| Variable | Value |
|----------|-------|
| `WEBHOOK_URL` | `https://<your-cloud-run-url>/webhook/email` |
| `WEBHOOK_SECRET` | The same value as `TARAN_WEBHOOK_SECRET` from Part 4 |

4. For `WEBHOOK_SECRET`, click **Encrypt** to store it securely.
5. Click **Save and Deploy**.

### Step 4: Set Up the Catch-All Routing Rule

This tells Cloudflare to send all incoming email to your worker.

1. Go back to **Email** > **Email Routing** > **Routing Rules**.
2. Find the **Catch-all** rule.
3. Click **Edit**.
4. Set the action to **Send to a Worker**.
5. Select your `mailbrief-email-forwarder` worker.
6. Click **Save**.

This means any email sent to `anything@yourdomain.com` will be processed by your worker, which forwards it to the backend. The backend then matches it to the correct user based on the "to" address.

---

## Part 9: Digest Scheduler (Google Cloud Scheduler)

The digest scheduler calls your backend every hour. The backend then checks if any users are due for a digest at that time (based on their preferred schedule and timezone) and generates one if needed.

### Step 1: Create the Scheduler Job

You already enabled the Cloud Scheduler API in Part 4. Now create the job:

```bash
gcloud scheduler jobs create http mailbrief-digest-trigger \
  --location us-central1 \
  --schedule "0 * * * *" \
  --uri "https://<your-cloud-run-url>/cron/digests" \
  --http-method POST \
  --headers "X-Webhook-Secret=<your-TARAN_WEBHOOK_SECRET>" \
  --attempt-deadline 120s
```

**What this does:**
- `--schedule "0 * * * *"` means "at minute 0 of every hour" (i.e., 1:00, 2:00, 3:00, etc.)
- `--uri` is the backend endpoint that triggers digest generation
- `--headers` sends the webhook secret so the backend accepts the request
- `--attempt-deadline 120s` gives the backend up to 2 minutes to respond (digest generation can take time for multiple users)

### Step 2: Verify the Scheduler

```bash
gcloud scheduler jobs list --location us-central1
```

You should see your `mailbrief-digest-trigger` job listed with status `ENABLED`.

### Step 3: Test It Manually

```bash
gcloud scheduler jobs run mailbrief-digest-trigger --location us-central1
```

This triggers the job immediately. Check the Cloud Run logs to confirm the backend received and processed the request:

```bash
gcloud run services logs read taran --region us-central1 --limit 20
```

Look for log entries related to digest generation. If there are no users with pending digests, it will simply log that no digests were needed.

---

## Part 10: Verification and Testing

Now that everything is deployed, walk through the entire flow to make sure it works end to end.

### Test 1: Website Loads

Open your browser and go to:
```
https://yourdomain.com
```

You should see the MailBrief login page with options to sign in with Google and GitHub.

**If the page does not load:** Check Part 6 (domain configuration). DNS changes can take up to an hour to propagate.

### Test 2: Sign In

1. Click **Sign in with Google** (or GitHub).
2. Complete the OAuth flow.
3. You should be redirected back to MailBrief and see the onboarding screen.

**If sign-in fails:** Check that the OAuth redirect URIs in Part 5 exactly match your domain, including `https://`.

### Test 3: Complete Onboarding

1. Follow the onboarding wizard.
2. Choose a username for your inbox (e.g., `yourname`).
3. After completing onboarding, you should see the main inbox view.

Your email address is now `yourname@yourdomain.com`.

### Test 4: Receive an Email

1. From a different email account (e.g., your personal Gmail), send an email to `yourname@yourdomain.com`.
2. Write a subject and a few sentences in the body.
3. Wait about 30 seconds.
4. Refresh the MailBrief inbox page.
5. You should see the email appear with an AI-generated summary.

**If the email does not appear:**
- Check Cloudflare Workers logs: **Workers & Pages** > your worker > **Logs**.
- Check Cloud Run logs: `gcloud run services logs read taran --region us-central1 --limit 20`
- Verify the webhook secret matches between the Cloudflare worker and the backend.

### Test 5: AI Extraction

Click on the received email. You should see:
- A summary of the email content.
- Extracted topics and key points.
- A reading time estimate.

**If extraction failed:** Check that your Anthropic API key is valid and has credits.

### Test 6: Trigger a Digest

```bash
curl -X POST "https://<your-cloud-run-url>/cron/digests" \
  -H "X-Webhook-Secret: <your-TARAN_WEBHOOK_SECRET>"
```

After a few seconds, go to the **Digests** page in MailBrief. You should see a generated digest summarizing your received emails.

**If no digest appears:** You may need at least a few emails in your inbox before a digest is generated.

### Test 7: Digest Email Delivery

If email sending is configured (Part 7), the digest should also arrive in your real email inbox. Check the email account you signed up with.

**If the email does not arrive:** Check your spam folder. Also check the Cloud Run logs for any Resend API errors.

---

## Part 11: Ongoing Maintenance

### Deploying Updates

**Frontend:** If you connected your GitHub repository to Vercel (the recommended approach), every push to the `main` branch automatically triggers a new deployment. Otherwise, run:
```bash
cd frontend
vercel --prod
```

**Backend:** Rebuild and deploy the container:
```bash
cd backend
gcloud run deploy taran \
  --source . \
  --region us-central1
```
This rebuilds the Docker image and deploys the new version. Existing environment variables are preserved.

### Monitoring

- **Backend logs:** `gcloud run services logs read taran --region us-central1 --limit 50`
- **Frontend logs:** Available at [vercel.com](https://vercel.com) > your project > **Deployments** > click a deployment > **Functions** tab
- **Health check:** `curl https://<your-cloud-run-url>/health`
- **Scheduler status:** `gcloud scheduler jobs list --location us-central1`
- **Email worker logs:** Cloudflare dashboard > **Workers & Pages** > your worker > **Logs**

### Estimated Costs

| Service | Free Tier | Estimated Monthly Cost |
|---------|-----------|----------------------|
| **Neon.tech** | 0.5 GB storage, 190 compute hours | Free for small usage |
| **Google Cloud Run** | 2 million requests/month, 360,000 GB-seconds | Free for small usage |
| **Google Cloud Scheduler** | 3 free jobs | Free |
| **Vercel** | Hobby plan (personal use) | Free |
| **Resend** | 3,000 emails/month | Free for small usage |
| **Anthropic** | Pay per token | ~$1-5/month for personal use |
| **Cloudflare** | Domain + Workers free tier | Domain registration cost only (~$10/year) |

For a single user or small team, the total monthly cost is typically under $5, with Anthropic API usage being the main variable cost.

### Scaling

- **Cloud Run** auto-scales the number of backend instances based on traffic.
- **Neon.tech** auto-scales storage as your database grows.
- **Vercel** auto-scales the frontend on the Hobby plan.
- **Resend** free tier supports 3,000 emails per month. Upgrade if you have many users.

### Backups

Neon.tech handles database backups automatically. The free tier includes 7 days of point-in-time recovery. For additional safety, you can manually export your database:

```bash
pg_dump "<your-neon-connection-string>" > mailbrief-backup.sql
```

---

## Part 12: Troubleshooting

### Email Not Appearing in Inbox

**Symptom:** You send an email to `yourname@yourdomain.com` but it never shows up in MailBrief.

**Check these in order:**

1. **Cloudflare MX records:** Go to Cloudflare > your domain > DNS. Verify the three MX records from Part 8 exist.
2. **Cloudflare email routing:** Go to Email > Email Routing. Verify the catch-all rule is set to your worker.
3. **Worker logs:** Go to Workers & Pages > your worker > Logs. Look for errors. Common issues:
   - `Webhook failed: 401` -- the webhook secret in the worker does not match `TARAN_WEBHOOK_SECRET` in the backend.
   - `Webhook failed: 500` -- the backend encountered an error. Check Cloud Run logs.
   - No log entries at all -- the email routing rule is not configured correctly.
4. **Cloud Run logs:** Run `gcloud run services logs read taran --region us-central1 --limit 20`. Look for errors during email processing.

### AI Extraction Failed

**Symptom:** The email appears but has no summary, topics, or key points.

1. **Check API key:** Verify `TARAN_ANTHROPIC_API_KEY` is set correctly in Cloud Run.
2. **Check credits:** Log in to [console.anthropic.com](https://console.anthropic.com) and verify your account has available credits.
3. **Check logs:** Run `gcloud run services logs read taran --region us-central1 --limit 20` and look for LLM-related errors.
4. **Model availability:** Ensure the model specified in `TARAN_ANTHROPIC_MODEL` is still available. Anthropic occasionally retires older model versions.

### Login Not Working

**Symptom:** Clicking "Sign in with Google" or "Sign in with GitHub" fails or redirects to an error page.

1. **Check redirect URIs:** The redirect URI registered with Google/GitHub must match exactly. Go back to Part 5 and verify:
   - Google: `https://yourdomain.com/api/auth/callback/google`
   - GitHub: `https://yourdomain.com/api/auth/callback/github`
2. **Check environment variables:** In Vercel > Settings > Environment Variables, verify `GOOGLE_CLIENT_ID`, `GOOGLE_CLIENT_SECRET`, `GITHUB_CLIENT_ID`, `GITHUB_CLIENT_SECRET`, and `BETTER_AUTH_SECRET` are all set.
3. **Check `BETTER_AUTH_URL`:** This must be `https://yourdomain.com` (with `https://`, without a trailing slash).
4. **Redeploy:** After changing environment variables in Vercel, you need to redeploy: `vercel --prod`.

### Digest Not Generated

**Symptom:** The scheduler runs but no digest appears.

1. **Check scheduler execution:** Run `gcloud scheduler jobs describe mailbrief-digest-trigger --location us-central1`. Look at `lastAttemptTime` and `status`.
2. **Check the cron header:** The scheduler must send `X-Webhook-Secret` in the headers. Verify this matches `TARAN_WEBHOOK_SECRET`.
3. **Check user settings:** Digests are only generated when it is the right time for a user based on their preferred schedule and timezone. Make sure your user account has a digest schedule configured.
4. **Check for emails:** Digests require at least some undigested emails. If your inbox is empty, no digest will be generated.
5. **Check logs:** `gcloud run services logs read taran --region us-central1 --limit 20` -- look for entries from the digest generator.

### Domain Not Loading

**Symptom:** Visiting `https://yourdomain.com` shows an error, timeout, or wrong page.

1. **DNS propagation:** DNS changes can take up to 48 hours (though usually much faster). Use [dnschecker.org](https://dnschecker.org) to verify your A record points to `76.76.21.21`.
2. **Cloudflare proxy:** Make sure the orange cloud is **off** (grey cloud / DNS only) for both the A record and the CNAME record. See Part 6.
3. **Vercel domain:** Go to Vercel > Settings > Domains. Your domain should show a green checkmark. If it shows "Invalid Configuration", the DNS records are not correct.
4. **SSL certificate:** If you see a browser security warning, wait 10 to 15 minutes for Vercel to provision the SSL certificate.

### Digest Email Not Delivered

**Symptom:** Digests appear in the app but the email notification never arrives.

1. **Check Resend:** Log in to [resend.com](https://resend.com) and check the **Emails** tab. Look for your digest emails -- they may show as "Delivered", "Bounced", or "Complained".
2. **Check spam:** Look in your spam/junk folder.
3. **DNS records:** Go to Resend > Domains and verify your domain shows "Verified". If not, recheck the DNS records from Part 7.
4. **API key:** Verify `TARAN_RESEND_API_KEY` is set in Cloud Run: `gcloud run services describe taran --region us-central1 --format='value(spec.template.spec.containers[0].env)'`
5. **SPF/DKIM:** Use [mail-tester.com](https://www.mail-tester.com) to send a test email and check your email authentication score.

### Backend Returns 500 Errors

**Symptom:** Various features fail and Cloud Run logs show HTTP 500 errors.

1. **Database connection:** Verify the Neon database is accessible. Go to [console.neon.tech](https://console.neon.tech) and check your project status. The free tier suspends after 5 minutes of inactivity -- the first request may be slow while it wakes up.
2. **Connection string:** Verify `TARAN_DB_URL` includes `?sslmode=require` at the end.
3. **Environment variables:** List all environment variables with `gcloud run services describe taran --region us-central1`. Verify nothing is missing.
4. **Logs:** The error message in the logs usually indicates the specific problem. Common ones:
   - `database unreachable` -- Neon connection issue
   - `context deadline exceeded` -- request took too long, may need to increase timeout
   - `unauthorized` -- API key mismatch between frontend and backend
