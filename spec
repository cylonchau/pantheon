%global linux_platform_   linux-amd64
%define linux_platform_   linux-amd64

Name: pantheon
Version: v7.3.1
Release: 1%{?dist}
Summary: An all-in-one observability solution which aims to combine the advantages of Prometheus and Grafana. It manages alert rules and visualizes metrics, logs, traces in a beautiful web UI.
Group: monitoring
License: GPL

%package ctl
Summary: 这是客户端升级时使用.
Group: n9e/cli
Vendor: cylonchau
Source0: pantheonctl
URL: https://github.com/cylonchau/pantheon/nightingale
BuildRoot: %{_tmppath}/%{name}-%{version}-buildroot

%description
Nightingale aims to combine the advantages of Prometheus and Grafana. It manages alert rules and visualizes metrics, logs, traces in a beautiful WebUI.

%description ctl
Nightingale commond line tool


%define __arch_install_post %{nil}
%define __os_install_post %{nil}
%global debug_package %{nil}

%prep

%install
rm -rf %{buildroot}
%{__install} -p -D %{SOURCE0} %{buildroot}/usr/sbin/pantheonctl


%files ctl
%defattr(-,root,root,-)
%attr(0555,root,n9e) /usr/sbin/pantheonctl

%changelog ctl
* Thu Aug 22 2024 root
- package cli