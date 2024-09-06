# nginx-subdomain-proxy example

More complex example where Nginx is set up to proxy all its requests to another reverse proxy, and only some
domains need to be secured.

Has full HTTP to HTTPS upgrades, some security configuration, and separate `server` blocks for the domains
that need Praga authentication vs those that do not. Should also support websocket connection upgrades.

The expectation is that the Nginx server IP is set as the A/AAAA record for multiple domains, and there is
another central load balancer or reverse-proxy setup after this Nginx install still.

You can adapt this rather easily if you want to define each subdomain separately to connect to different
backend servers or host local files, or whatever else.
