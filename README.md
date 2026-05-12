# red

![red](https://user-images.githubusercontent.com/141232/54882450-bb85b200-4e8c-11e9-8bd9-37cf43b5b1ed.gif)

_Red_ is a terminal log analysis tools.

## Usage

Pipe JSON stream logs into _red_ and specify a few fields to display. For example using with kubernetes:

```bash
kubectl logs ... | red level message
```

You will see combined logs with trend sparkline and total count.

## Install

```bash
go install github.com/antonmedv/red@latest
```

## Usage

Pipe newline-delimited JSON into `red` and list fields to group by:

```bash
... | red <field> [<field> ...]
```

Keys:

- `↑` / `↓` — select a row
- `Enter` — open detail view for the selected row
- `Esc` — close detail view

Flags:

- `-trend <duration>` — trend window (default `10s`)
- `-distance <n>` — Levenshtein distance for grouping (default `3`)

## Kubernetes

```bash
kubectl logs -f deploy/api | red level message
```

## Caddy access logs

Caddy's default access log is JSON, so it works directly. Most useful fields
are nested under `request`, so flatten them with `jq` first.

Enable JSON access logs in your Caddyfile:

```caddyfile
example.com {
    log {
        output file /var/log/caddy/access.log
        format json
    }
    reverse_proxy localhost:8080
}
```

Then watch live traffic:

```bash
tail -F /var/log/caddy/access.log \
  | jq -c '{status, method: .request.method, uri: .request.uri, ip: .request.client_ip}' \
  | red status method uri
```

### Visitor stats recipes

Top requested paths:

```bash
tail -F /var/log/caddy/access.log \
  | jq -c '{uri: .request.uri}' \
  | red uri
```

Top client IPs (visitors):

```bash
tail -F /var/log/caddy/access.log \
  | jq -c '{ip: .request.client_ip}' \
  | red ip
```

Status code distribution:

```bash
tail -F /var/log/caddy/access.log \
  | jq -c '{status}' \
  | red status
```

Top user agents (with longer trend window):

```bash
tail -F /var/log/caddy/access.log \
  | jq -c '{ua: (.request.headers."User-Agent"[0] // "-")}' \
  | red -trend 1m ua
```

## Nginx access logs

Nginx's default log format is plain text. Convert to JSON first — either by
configuring nginx to emit JSON, or by piping through a parser.

Configure nginx for JSON access logs in `nginx.conf`:

```nginx
log_format json_combined escape=json
  '{'
    '"time":"$time_iso8601",'
    '"remote_addr":"$remote_addr",'
    '"method":"$request_method",'
    '"uri":"$request_uri",'
    '"status":$status,'
    '"bytes_sent":$bytes_sent,'
    '"referer":"$http_referer",'
    '"user_agent":"$http_user_agent"'
  '}';

access_log /var/log/nginx/access.log json_combined;
```

Then:

```bash
tail -F /var/log/nginx/access.log | red status method uri
```

Top visitor IPs:

```bash
tail -F /var/log/nginx/access.log \
  | jq -c '{remote_addr}' \
  | red remote_addr
```

## Ideas for visitors analytics UI

Now red show one table only. For watching site visitors in live mode, more
nice UI can help very much. Here is some simple ideas — written by simple
words, without difficult terms.

### Big numbers on top

Show some big numbers on top of screen, so you can see most important
informations from one look:

- **Visitors now** — how many peoples was on site in last 5 minutes
- **Page views per minute** — how much site is busy
- **Errors per minute** — how many 4xx and 5xx answers from server
- **Middle page time** — how fast pages is loading

This numbers change in live time when new logs is coming.

### Divide screen on panels

Instead of one table, show many small panels near each other:

```
┌─────────────────────┬──────────────────────┐
│  Top pages          │  Top countries       │
│  /                  │  USA                 │
│  /blog              │  Great Britain       │
│  /pricing           │  Germany             │
├─────────────────────┼──────────────────────┤
│  Status codes       │  Visitors now: 42    │
│  ▇▇▇▇▇ 200 (89%)    │  Page views: 1.2k/m  │
│  ▇ 404 (8%)         │  Errors: 3/m         │
│  ▏ 500 (3%)         │  Middle load: 240ms  │
└─────────────────────┴──────────────────────┘
```

Every panel watch one thing. You see all picture without scrolling.

### Colors with sense

Use colors, so problems is visible:

- **Green** — all good (200 OK)
- **Yellow** — be careful (404 Not Found, slow pages)
- **Red** — bad (500 errors, very very slow pages)

When you see many red — something is broken. When all green — everything
works in normal mode.

### Map of world for visitors

Draw small ASCII map of world. Make light countries where visitors are
sitting right now. More bright color means more peoples from this country.

For this need IP-to-country base (free MaxMind GeoLite2 is good and works
without problems).

### Click for go deeper

Click on row in "Top pages" — and other panels make filter only for this
page. Click on country — see visitors only from there. Press Esc for come
back.

So red become not only viewer, but tool for research.

### Pause and back in time

Pause button for stop screen. Very comfortable when something interesting
happen and you want look on it without new logs running up.

Small time slider for scroll back on last few minutes.

### Find strange traffic

Mark rows what look not normal:

- Same IP open many pages very fast (maybe is bot)
- Many 404 from one IP (somebody search secret files)
- Sudden big quantity of one user agent
- New country what never was before

Show small warning icon near row. Click on it for see why.

### Browsers and devices

Panel what show what peoples use:

```
Browsers           Devices
Chrome   ▇▇▇▇ 62%   Mobile  ▇▇▇ 48%
Safari   ▇▇ 24%     Desktop ▇▇▇ 45%
Firefox  ▇ 10%      Tablet  ▏ 7%
Other    ▏ 4%
```

This data is parsing from User-Agent header.

### Search line

Search line on top. Write path, IP or status code, and all UI make filter
for this. Press Esc for clean.

### Why this is important

Now red is good for engineers what read logs. With this ideas it can also
help to not-engineers — product peoples, marketing, support — for see what
is happening on site, and not learn terminal commands.

## License

MIT
