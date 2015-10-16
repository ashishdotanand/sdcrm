# -*- mode: ruby -*-
# vi: set ft=ruby :
# Vagrantfile API/syntax version. Don't touch unless you know what you're doing!
VAGRANTFILE_API_VERSION = "2"

Vagrant.configure(VAGRANTFILE_API_VERSION) do |config|
  config.vm.box = "ubuntu/trusty32"  #More boxes can be discovered at https://vagrantcloud.com/ubuntu
  config.vm.network :forwarded_port, guest: 80, host: 8080, auto_correct: true
  config.vm.network :forwarded_port, guest: 8080, host: 9000, auto_correct: true
  config.vm.network :forwarded_port, guest: 8000, host: 7000, auto_correct: true
  config.vm.network :forwarded_port, guest: 5432, host: 15432, auto_correct: true
  config.vm.provision "shell", path: "./Vagrant-setup/bootstrap.sh"
end
