# Simple Beacon-of-Beacons (BoB)

This is a simple implementation of a beacon-of-beacons: a service that
exposes the GA4GH beacon API and forwards the request on to a list of
other beacon systems. This implementation is intended to serve as a
test bench for experimenting with new security and federation
features. This is a work-in-progress, so caveat emptor.


## Authenticated Beacon of Beacons Service

The beacon-of-beacons service (BoB) allows a user to query multiple
GA4GH beacons at one time with a single query. This implementation
demonstrates the use of authentication and authorization for a
beacon-of-beacons.

When you access the main page of the web service, you will be
presented with a selection of identity providers at which to
authenticate. In this implementation, the set of providers is
configured statically before the BoB is started -- see "Configuration"
below.

Upon authentication to an identity provider, information about the
authentication is provider to the BoB in the form of two tokens: (a)
an OAUTH 2.0 access token access token and (b) an OpenID Connect ID
token. The ID token is a secure object that allows a third party to
verify that the indentity provider successfully authenticated a given
principal. The access token allows the BoB to obtain more detailed
information about the principal, including information such as a
human-readable name, email address, and other attributes recorded
about the principal. At present, the BoB only uses this information to
display the human-readable name of the authenticated principal.

When you make a beacon query to the BoB, the BoB forwards the query to
a statically-configured set of beacons along with the two tokens
described above. Each beacon can verify the ID token and then use the
access token to obtain further information about the principal. A
beacon then uses this information to make an authorization decision,
allowing or denying the beacon request and returning whatever
information it deems appropriate. 

This decentralized authorization model allows each beacon to determine
and enforce its own criteria. For example, a beacon may wish to
provide service only to principals who have a validated email from an
academic institution. Similarly, a beacon may wish to base its
authorization decision on the presence or absence of specific custom
claims from the identity provider, e.g. a claim asserting that the
principal is deemed a bona fide researcher, etc.

## BoB Query Flow

There are seven endpoints exposed by the BoB service, and in a typical
interaction, they are invoked in roughly this order:

1. `/static/` is the endpoint for fetching javascript, css, and html
templates.

2. `/login` presents a choice of identity providers for
authentication. On selecting one, redirect to the next endpoint...

3. `/login/<provider>` redirects the browser to the selected provider.
This step is necessary because the BoB needs to keep a record of the
authentication request so that it can correlate the request with the
callback from the identity provider, which is delivered to...

4. `/callback` receives the authentication credentials (access and ID
tokens) from the identity provider. This endpoint looks up the login
request record and from it, determines the original page requested.
This page is almost always...

5. `/` the main query page, which allows the user to enter a beacon
query and send it around to all of the configured beacons. As noted
above, the access and ID tokens are also sent along. This is done by
adding HTTP headers for the tokens. The access token is sent in the
`Authorization` header:

  ```
  Authorization: Bearer <access_token>
  ```

  and the ID token is sent in a non-standard header:

  ```
  IDToken: <id_token>
  ```

  The queries to the individual beacons are performed in parallel. As
  the results come back for each beacon, they are sent over to the
  browser using a websocket...

6. `/ws` the websocket endpoint used to actually post the query. The
websocket conenction is started when the main query page is loaded,
and the beacon query is sent to the BoB server asynchronously. The
responses are also delivered over this websocket channel
asynchronously.

7. `/logout` used to terminate the session and log out.

## Configuration


### Command-line Switches
There are a number of command-line switches that can be used to
configure the server:

```
Usage of ./bob:
  -config string
        Configuration directory (default "./config")
  -host string
        Host name (default "127.0.0.1")
  -port int
        Port on which to run server (default 8080)
  -timeout int
        Timeout for beacon queries, in seconds (default 20)
```

In addition, there are two sets of resources that must be statically
configured: the set of identity provider and the set of beacons.

### Identity Provider Configuration

To configure the set of identity providers, place one configuration
file for each provider in the `/config/idp` directory. Keep in mind
that the default location of the config directory may be different
based on the -config command line switch.

An identity provider file looks like this:

```
{
    "name": "Genecloud Test IdP",
    "endpoint": "https://login.dev.genecloud.com",
    "clientIdEnv": "GENECLOUD_CLIENT_ID",
    "clientSecretEnv": "GENECLOUD_CLIENT_SECRET",
    "redirectURL": "http://127.0.0.1:8080/callback"
}
```

1. `name` is an arbitrary human-readable name

2. `endpoint` is the main URL of the identity provider, 

3. `clientIdEnv` specifies the name of an environment variable that
holds the unique client ID assigned by the identity provider.
Alternatively, you can provide a field `clientId` that directly
contains the client id. The environment variable work-around is there
so that you do not need to commit any actual client secrets into a
code repository.

4. `clientSecretEnv` is an environment variable that holds the client
secret established with the identity provider. Alternatively, specify
`clientSecret`.

5. `redirectURL` is the URL to which the user should be returned upon
authentication to the identity provider.



## To do
* Extract common code from beacon versions
* Automated tests
* Cleanup and documentation
* Dockerfile
* Icons, other metadata for IDPs
* Improve UI for login
* Redirect URL -- is it even used? Does the service just use it's own?
