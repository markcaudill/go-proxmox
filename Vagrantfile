# -*- mode: ruby -*-
# vi: set ft=ruby :

$script = <<-SCRIPT
# Provision per https://pve.proxmox.com/wiki/Install_Proxmox_VE_on_Debian_Buster
# Adapted from https://github.com/Telmate/proxmox-api-go/blob/master/Vagrantfile

# make sure hostname can be resolved via /etc/hosts
sed -i "/127.0.1.1/d" /etc/hosts
PVE_IP=$(hostname --all-ip-address | awk '{print $1}')
if [ -z "$(grep $PVE_IP /etc/hosts)" ]; then
	echo "$PVE_IP $(hostname)" > /etc/hosts
fi

# add proxmox repository and its key
echo "deb http://download.proxmox.com/debian/pve buster pve-no-subscription" > /etc/apt/sources.list.d/pve-install-repo.list
wget http://download.proxmox.com/debian/proxmox-ve-release-6.x.gpg -O /etc/apt/trusted.gpg.d/proxmox-ve-release-6.x.gpg
chmod +r /etc/apt/trusted.gpg.d/proxmox-ve-release-6.x.gpg  # optional, if you have a non-default umask

# update repositories and system
apt update && apt full-upgrade -y

# install proxmox packages
apt install -y proxmox-ve postfix open-iscsi

# don't scan for other operating systems
apt-get remove -y os-prober

# set root password so that we can use it to login to Proxmox API
sudo -i passwd <<EOF
root
root
EOF
SCRIPT


Vagrant.configure("2") do |config|
  config.vm.box = "debian/contrib-buster64"
  config.vm.network "forwarded_port", guest: 8006, host: 8006, host_ip: "127.0.0.1"
  config.vm.synced_folder ".", "/vagrant", disabled: true

  config.vm.provider "virtualbox" do |vb|
	vb.cpus = 2
    vb.memory = 2048
  end

  config.vm.provision "shell",
	  privileged: true,
	  env: { "DEBIAN_FRONTEND" => "noninteractive" },
	  inline: $script
end
