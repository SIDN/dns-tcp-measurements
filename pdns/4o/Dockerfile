FROM auth-powerdns:latest

# Copy configuration files and zone file
COPY pdns.conf /var/lib/powerdns/pdns.conf
COPY named.conf /var/lib/powerdns/named.conf
COPY 4o.zone.signed /var/lib/powerdns/zones/4o.zone.signed

# Make sure PowerDNS can find the configuration files
RUN ln -s /var/lib/powerdns/pdns.conf /usr/local/etc/pdns.conf 

# Create empty database for the dnssec related data
RUN pdnsutil create-bind-db /var/lib/powerdns/bind-dnssec-db.sqlite3

# Sign the zone     
RUN pdnsutil set-presigned 4o
RUN pdnsutil set-nsec3 4o
