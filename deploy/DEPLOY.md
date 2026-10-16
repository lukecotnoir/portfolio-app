# going live — lukecotnoir.dev over ssh

Goal: `ssh lukecotnoir.dev` drops anyone into the portfolio TUI,
while `https://lcotnoir.dev` serves the Angular build.

## what you need (one-time, ~$10/yr + $5/mo)

- a domain — e.g. `lcotnoir.dev` (Namecheap/Porkbun/Cloudflare, ~$10-15/yr)
- a small VPS — Hetzner CX11 / DigitalOcean droplet / OVH / Oracle free tier (~$5/mo)

## DNS

Two records, pointing at the VPS:

    A     lukecotnoir.dev      → <vps ip>
    A     www.lukecotnoir.dev  → <vps ip>   (optional)

## the catch: sshd already owns port 22

Your VPS runs OpenSSH for admin on :22. The portfolio needs :22 so
guests can type `ssh lukecotnoir.dev` with no `-p` flag. Fix: move
admin sshd to a different port, give :22 to the app.

```sh
# on the VPS — move admin sshd to 2222 FIRST (don't lock yourself out)
sudo sed -i 's/^#\?Port .*/Port 2222/' /etc/ssh/sshd_config
sudo systemctl reload sshd
# verify you can still get in before continuing:
ssh -p 2222 you@<vps ip>
```

## install the app

```sh
# build a static linux binary locally, then ship it
cd ssh && GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o lcotnoir .

scp -P 2222 lcotnoir content.json you@<vps ip>:/tmp/

# on the VPS
sudo useradd -r -s /usr/sbin/nologin lcotnoir
sudo mkdir -p /opt/lcotnoir/.ssh
sudo mv /tmp/lcotnoir /tmp/content.json /opt/lcotnoir/
sudo cp deploy/lcotnoir-ssh.service /etc/systemd/system/
sudo chown -R lcotnoir:lcotnoir /opt/lcotnoir

sudo systemctl daemon-reload
sudo systemctl enable --now lcotnoir-ssh
```

The unit grants only `CAP_NET_BIND_SERVICE` — the app binds :22
without root and guests land in the TUI, never a shell. There is no
login user and no shell behind it; the TUI *is* the endpoint.

## firewall

```sh
sudo ufw allow 22/tcp    # guests — the portfolio
sudo ufw allow 2222/tcp  # you — admin sshd
sudo ufw enable
```

## https (for the web build)

GitHub Pages serves the site; point your domain at Pages
(CNAME `lcotnoir.dev` → `you.github.io`) or use Cloudflare in front.
Pages handles the cert.

## test

```sh
ssh lukecotnoir.dev      # should land in the TUI — type 'help', then 'exit'
ssh -p 2222 <vps ip>     # your admin access, unchanged
```

## ops notes

- `journalctl -u lcotnoir-ssh -f` to watch it work
- the binary embeds content.json — rebuilding after a resume update is
  `make ssh` + `scp` + `systemctl restart lcotnoir-ssh`
- anonymous guest access is read-only by design; there is nothing to
  rate-limit except your own curiosity
