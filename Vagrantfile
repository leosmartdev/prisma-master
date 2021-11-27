# -*- mode: ruby -*-
# vi: set ft=ruby :

# Vagrantfile API/syntax version. Don't touch unless you know what you're doing!
VAGRANTFILE_API_VERSION = "2"

Vagrant.configure(VAGRANTFILE_API_VERSION) do |config|
  config.vm.box = "ubuntu/xenial64"
  config.vm.hostname = "tms-dev"

  config.ssh.forward_x11 = true

  config.vm.define "cc1", primary: true do |cc1|
    cc1.vm.network "forwarded_port", guest: 8080, host: 8080
    cc1.vm.network "forwarded_port", guest: 8000, host: 8000
    cc1.vm.network "forwarded_port", guest: 8237, host: 8237
    cc1.vm.network "forwarded_port", guest: 8080, host: 8080
    cc1.vm.network "forwarded_port", guest: 8081, host: 8081
    cc1.vm.network "forwarded_port", guest: 8089, host: 8089
    cc1.vm.network "forwarded_port", guest: 8181, host: 8181
    cc1.vm.network "forwarded_port", guest: 9090, host: 9090
    cc1.vm.network "forwarded_port", guest: 9099, host: 9099
    cc1.vm.hostname = "cc1"
    cc1.vm.network "private_network", ip: "110.0.0.10"
    cc1.vm.network "private_network", ip: "111.0.0.10"
    cc1.vm.network "private_network", ip: "112.0.0.10"
    cc1.vm.network "private_network", ip: "113.0.0.10"
    cc1.vm.network "private_network", ip: "210.0.0.10"
    cc1.vm.provision :shell, path: "vagrant/bootstrap"
    cc1.vm.provider "virtualbox" do |vb|
        vb.name = "tms-dev-cc1"
        vb.memory = 4096
        vb.cpus = "2"
        vb.customize ["modifyvm", :id, "--nictype1", "virtio"]
    end
  end

  config.vm.define "u18" do |u18|
    u18.vm.box = "ubuntu/bionic64"

    u18.vm.network "forwarded_port", guest: 8080, host: 8080
    u18.vm.network "forwarded_port", guest: 8000, host: 8000
    u18.vm.network "forwarded_port", guest: 8237, host: 8237
    u18.vm.network "forwarded_port", guest: 8080, host: 8080
    u18.vm.network "forwarded_port", guest: 8081, host: 8081
    u18.vm.network "forwarded_port", guest: 8089, host: 8089
    u18.vm.network "forwarded_port", guest: 8181, host: 8181
    u18.vm.network "forwarded_port", guest: 9090, host: 9090
    u18.vm.network "forwarded_port", guest: 9099, host: 9099
    u18.vm.hostname = "u18"
    u18.vm.network "private_network", ip: "110.0.0.10"
    u18.vm.network "private_network", ip: "111.0.0.10"
    u18.vm.network "private_network", ip: "112.0.0.10"
    u18.vm.network "private_network", ip: "113.0.0.10"
    u18.vm.network "private_network", ip: "210.0.0.10"
    u18.vm.provision :shell, path: "vagrant/bootstrap"
    u18.vm.provider "virtualbox" do |vb|
        vb.name = "tms-dev-cc118"
        vb.memory = 1536
        vb.cpus = "1"
        vb.customize ["modifyvm", :id, "--nictype1", "virtio"]
    end
  end
  config.vm.define "u18", autostart: false

  config.vm.define "cc2" do |cc2|
    cc2.vm.network "forwarded_port", guest: 8237, host: 18237
    cc2.vm.network "forwarded_port", guest: 8080, host: 18080
    cc2.vm.network "forwarded_port", guest: 8081, host: 18081
    cc2.vm.network "forwarded_port", guest: 8089, host: 18089
    cc2.vm.network "forwarded_port", guest: 8181, host: 18181
    cc2.vm.network "forwarded_port", guest: 9090, host: 19090
    cc2.vm.network "forwarded_port", guest: 9099, host: 19099
    cc2.vm.hostname = "cc2"
    cc2.vm.network "private_network", ip: "110.0.0.11"
    cc2.vm.network "private_network", ip: "111.0.0.11"
    cc2.vm.network "private_network", ip: "112.0.0.11"
    cc2.vm.network "private_network", ip: "113.0.0.11"
    cc2.vm.provision :shell, path: "vagrant/bootstrap"
    cc2.vm.provider "virtualbox" do |vb|
        vb.name = "tms-dev-cc2"
        vb.memory = 4096
        vb.cpus = "2"
        vb.customize ["modifyvm", :id, "--nictype1", "virtio"]
    end
   end
   config.vm.define "cc2", autostart: false

   config.vm.define "tmsdemo" do |demo|
       demo.vm.network "forwarded_port", guest: 8237, host: 28237
       demo.vm.network "forwarded_port", guest: 8080, host: 28080
       demo.vm.network "forwarded_port", guest: 8081, host: 28081
       demo.vm.network "forwarded_port", guest: 8089, host: 28089
       demo.vm.network "forwarded_port", guest: 8181, host: 28181
       demo.vm.network "forwarded_port", guest: 9090, host: 29090
       demo.vm.network "forwarded_port", guest: 9099, host: 29099
       demo.vm.hostname = "tmsdemo"
       demo.vm.network "private_network", ip: "110.0.0.12"
       demo.vm.network "private_network", ip: "111.0.0.12"
       demo.vm.network "private_network", ip: "112.0.0.12"
       demo.vm.network "private_network", ip: "113.0.0.12"
       demo.vm.synced_folder "./dist/", "/dist"
       demo.vm.provider "virtualbox" do |vb|
           vb.name = "tms-dev-demo"
           vb.memory = 4096
           vb.cpus = "2"
           vb.customize ["modifyvm", :id, "--nictype1", "virtio"]
       end
      end
      config.vm.define "tmsdemo", autostart: false
end

