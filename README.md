linux-webui
===========

A simple web control panel for linux servers

This project has been developed and tested on Gentoo, so that is the one distro where I can safely say that it works flawlessly. I plan to test it on other distros such as Debian (maybe Ubuntu too), Arch and Fedora (maybe CentOS too) within the next couple of weeks as time permits (I am a student). If you try to run it on any of the untested distros, it will have a few bugs or may not work at all in some cases. If you want me to specifically test it on a particular distro, please add an issue to this project and I will try to expedite the testing for that distro (as long as the distro can be easily installed in a virtual machine).

Requirements
  - Apache
  - PHP
  - sudo root access for apache user (see below)

Installation via Git (Recommended)
  - Install Apache (guides for how to do this are available online for almost every linux distro).
  - Install PHP (guides for how to do this are available online for almost every linux distro).
  - Clone the git repository into your web server directory (usually /var/www or its subdirectories) by running 'git clone https://github.com/virajchitnis/linux-webui.git'.
  - Give sudo root access to Apache user by adding 'apache ALL=(ALL) NOPASSWD: ALL' to the /etc/sudoers file where 'apache' is the Apache user (apache on Gentoo, www-data on Debian, httpd on CentOS, etc).
  - Future updates to the web application can be installed by simply clicking the 'Update' button in the about page of the project.

Installation via tar.gz (Not recommended)
  - Install Apache.
  - Install PHP.
  - Download the tar.gz for the latest release version from GitHub into your web server directory by running 'wget https://github.com/virajchitnis/linux-webui/archive/v1.1.1.tar.gz' (replace v1.1.1 with the version you wish to download).
  - Extract by running 'tar -zxvf v1.1.1.tar.gz'.
  - Delete the downloaded tar.gz by running 'rm v1.1.1.tar.gz'
  - Future updates to the web application will have to be installed by downloading the tar.gz for the newer version of the application and then extracting it over the current version's folder.
