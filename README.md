linux-webui (beta)
==================
A simple web control panel for linux servers

This project has been developed and tested on Gentoo. Other distros will be tested as and when I have time (I am a student). Most Linux distros are quite similar to each other, so a most of the control panel should be functional on any distro. If you manage to get it working on some other distro, please let me know so that I can update the OS support info.

If you install the linux-webui on a web facing server, it is very important that you **prevent unauthorized access** to it. This can be done using the security features provided by your web server. If you are using Apache, you would enable either '*mod_auth_basic*' or '*mod_auth_digest*'. Documentation and tutorials on how to enable this can be found in many places online.

OS support
----------

* Gentoo

Requirements
------------

* Apache 2.2+
* PHP 5+
* sudo root access for apache user (see below)

Installation via Git (Recommended)
----------------------------------

1. Install Apache
2. Install PHP
3. Clone the git repository into your web server directory (usually /var/www or its subdirectories) by running 'git clone https://github.com/virajchitnis/linux-webui.git'.
4. Give sudo root access to Apache user by adding 'apache ALL=(ALL) NOPASSWD: ALL' to the /etc/sudoers file where 'apache' is the Apache user (apache on Gentoo, www-data on Debian, httpd on CentOS, etc).
5. Future updates to the web application can be installed by simply clicking the 'Update' button in the about page of the project.

Installation via tar.gz (Not recommended)
-----------------------------------------

1. Install Apache
2. Install PHP
3. Download the tar.gz for the latest release version from GitHub into your web server directory by running 'wget https://github.com/virajchitnis/linux-webui/archive/v1.1.2.tar.gz' (replace v1.1.2 with the version you wish to download).
4. Extract by running 'tar -zxvf v1.1.2.tar.gz'.
5. Delete the downloaded tar.gz by running 'rm v1.1.2.tar.gz'
6. Future updates to the web application will have to be installed by downloading the tar.gz for the newer version of the application and then extracting it over the current version's folder.
