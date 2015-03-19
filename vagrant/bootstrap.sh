#!/bin/bash

cat /vagrant/vagrant/packages | xargs yum install -y

service mysqld restart

mysql -u root -e"CREATE DATABASE graphite_web;"
mysql -u root -e"GRANT ALL PRIVILEGES ON graphite_web.* TO 'graphite-web'@'localhost' identified by 'password';"

cat << EOF >>/etc/graphite-web/local_settings.py
DATABASES = {
    'default': {
        'NAME': 'graphite_web',
        'ENGINE': 'django.db.backends.mysql',
        'USER': 'graphite-web',
        'PASSWORD': 'password',
        'HOST': 'localhost',
        'PORT': '3306'
    }
}
EOF

# This step sets up the db for django. This consumes the initial_data.json file.
# If you need to regenerate the initial_data.json file, it can be done thusly:
# python /usr/lib/python2.6/site-packages/graphite/manage.py dumpdata | python -mjson.tool >> /vagrant/vagrant/initial_data.json
cd /usr/lib/python2.6/site-packages/graphite/
cp /vagrant/vagrant/initial_data.json .
python manage.py syncdb --noinput

# The default graphite-web.conf file was missing an alias for /content causing it to not serve up static files.
rm -rf /etc/httpd/conf.d/graphite-web.conf
ln -s /vagrant/vagrant/graphite-web.conf /etc/httpd/conf.d/graphite-web.conf

# Adjust default storage schema for carbon
sed -i 's/60s:1d/60s:30d/' /etc/carbon/storage-schemas.conf

service carbon-cache restart
service httpd restart

chkconfig mysqld on
chkconfig carbon-cache on
chkconfig httpd on

# disalbe the firewall
service iptables stop
chkconfig iptables off

echo "export GOPATH=/vagrant" >> /etc/bashrc
