# Cert

A simple [Let's Encrypt](https://letsencrypt.org/) client.

# Installation

    $ git clone https://github.com/caldwell/cert
    $ cd cert
    $ bundle install

## Advanced setup

I keep my config file in a separate directory so it can be revision
controlled independently of the main code. I put a symlink to the main cert
code in that directory (called cert-src) and wrote a little script that
launches cert:

    #!/bin/bash

    confdir=$(dirname $0)
    cd "$confdir"
    confdir=$(pwd)
    cd "$confdir/cert-src"
    bundle exec -- cert -c "$confdir" "$@"

Then I set up a cron job that looks like this:

    37 0 * * * cd /my/ssl/dir && ./cert -vv

# Configuration

Cert is configured by a single yaml file (named `config.yaml`). Here's an example:

    cyrus-imap-server:
      section   : "IMAP Server"
      cn        : "mail.example.com"
      user      : cyrus
      group     : mail
      on_renew    : "systemctl restart cyrus-imapd.service"

    smtp.example.com:
      section     : "Outbound Mail Server"
      cn          : "smtp.example.com"
      well-known  : root@remote-server.example.com:/var/www/html/.well-known/acme-challenge
      scp         : root@remote-server.example.com:/etc/ssl/private/
      on_renew    : "ssh root@remote-server.example.com systemctl restart postfix.service"

    jabber.example.com:
      section       : "Jabber Server"
      cn            : "example.com"
      group         : ejabberd
      combined      : true
      on_renew      : "systemctl restart ejabberd.service"
      alt:
        - DNS: jabber.example.com
        - DNS: example.com

This file will create the following certs:

    certs/2018/cyrus-imap-server-2018-03-05.cert.pem
    certs/2018/cyrus-imap-server-2018-03-05.csr.pem
    certs/2018/cyrus-imap-server-2018-03-05.key.pem
    certs/2018/smtp.example.com-2018-03-05.cert.pem
    certs/2018/smtp.example.com-2018-03-05.csr.pem
    certs/2018/smtp.example.com-2018-03-05.key.pem
    certs/2018/jabber.example.com-2018-03-05.cert.pem
    certs/2018/jabber.example.com-2018-03-05.combined.pem
    certs/2018/jabber.example.com-2018-03-05.csr.pem
    certs/2018/jabber.example.com-2018-03-05.key.pem

Those aren't super helpful, so it also creates a canonically named symlinks
that point to the latest files:

    certs/cyrus-imap-server.cert.pem
    certs/cyrus-imap-server.csr.pem
    certs/cyrus-imap-server.key.pem
    certs/smtp.example.com.cert.pem
    certs/smtp.example.com.csr.pem
    certs/smtp.example.com.key.pem
    certs/jabber.example.com.cert.pem
    certs/jabber.example.com.combined.pem
    certs/jabber.example.com.csr.pem
    certs/jabber.example.com.key.pem

## Configuration details

Certificates are defined by top level hash entries in the `config.yaml`
file. The key is what defines the name of the resulting files on the
disk. This means they are restricted to the character set that is available
for files (IE, no `/`, though space will probably work but hasn't been
tested).

Each cert entry is also a hash where the keys define various options:

  - `cn` (Required): The "Common Name" for the certificate. This is the main
    domain name, fully qualified.

  - `country`, `organization`, `section`, `state`, `locality`, `email`
    (Not Required): These go into the various fields of the certificate
    "Subject". These are all ignored by Let's Encrypt, but you can use them as
    documentation.

  - `alt` (Not Required): This is an array of hashes of alternate
    names. Currently, the only support key of these hashes is 'DNS', and the
    value should be a fully qualified DNS name.

  - `user` (Not Required): Set the user of the key file. This requires cert
    to be run as root. It's usually better to set the group instead.

  - `group` (Not Required): Set the group of the key file. Use this when the
    server that needs to read the key file doesn't have the correct
    permissions by default. This requires cert to be run as root, or for the
    user running cert to have membership in the group.

  - `well-known` (Not Required, Default: `/var/www/well-known/acme-challenge/`):
    The location of the well-known acme-challenge directory. This can either
    be a directory name on the local machine, or an SSH style
    "[user@]machine:directory" location.

    Setting up the directory to respond to HTTP requests for
    `/well-known/acme-challenge` is not the purview of this script and must
    be done manually.

  - `combined` (Not Required, Default: false): If true, creates a
    "certs/<name>.combined.pem" which includes both the cert and the
    key. This file is created with the same permissions as the key.

  - `scp` (Not Required): If set, will use `scp` to copy the certificate
    file and key to the specified remote directory.

  - `on_renew` (Not Required): If set, will run the specified
    command. Usually this is for restarting or reloading the daemon to get
    it to start using the new certificate.

There are also some top level keys to set global options:

  - `RSA_BITS` (Not Required, Default 4096): The number of bits in created
    RSA private keys.

  - `include` (Not Required): An array of files to include. These YAML files
    are loaded and merged into the config file. Note, including only works
    in the main config files. Included files cannot include more files.

# Invocation

Running cert will check the certificates, acquire any new certificates, and
renew any that are within 10 days of expiry.

## Command line options

 - `-v`, `--verbose`: Turn up the verbosity. By default the program is
   silent except for errors. Add the option multiple times to increase the
   logging detail.

 - `-n`, `--dry-run`: Don't really do anything (not super useful unless you
   have `--verbose` on as well).

 - `-c`, `--config=<confdir>`: Use a different config directory. The config
   directory is where the `config.yaml` file and the `certs` output
   directory are located. By default the config directory is in the same
   directory as the cert program itself.

 - `--staging`: Use the Let's Encrypt staging CA--this will go through the
   motions but the certs you recieve will not be usable. Use this to test
   your configuration, so you don't use up your Let's Encyrpt certificate
   quotas on tests.

# Bugs

There are no bugs in this program. 😇 
If you disagree, you can file bug
report on the github page: https://github.com/caldwell/cert

# Author

Copyright 2016-2020 David Caldwell <david@porkrind.org>

This program is free software: you can redistribute it and/or modify it
under the terms of the GNU General Public License as published by the Free
Software Foundation, either version 3 of the License, or (at your option)
any later version.

This program is distributed in the hope that it will be useful, but WITHOUT
ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or
FITNESS FOR A PARTICULAR PURPOSE.  See the GNU General Public License for
more details.

You should have received a copy of the GNU General Public License along with
this program.  If not, see <http://www.gnu.org/licenses/>.
