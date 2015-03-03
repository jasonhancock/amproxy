# -*- mode: ruby -*-
# vi: set ft=ruby :

# Vagrantfile API/syntax version. Don't touch unless you know what you're doing!
VAGRANTFILE_API_VERSION = "2"

Vagrant.configure(VAGRANTFILE_API_VERSION) do |config|
  config.vm.box = "jsnby_centos6_x86_64"

  # CentOS 6.4 x86_64 with EPEL
  config.vm.box_url = "https://dl.dropboxusercontent.com/sh/vomjerdguovpno9/3k5cO6YA6w/centos6_x86_64.box"

  config.vm.network :forwarded_port, host: 8081, guest: 80
  config.vm.provision :shell, :path => "vagrant/bootstrap.sh"
end
