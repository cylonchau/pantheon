Name: pantheon
Version: ${VERSION}
Release: 1%{?dist}
Summary: pantheon
Group: pantheon
License: Apache License 2.0

URL: https://github.com/cylonchau/pantheon
Name: pantheon
Version: ${VERSION}
Release: 1%{?dist}
Summary: An all-in-one observability solution which aims to combine the advantages of Prometheus and Grafana. It manages alert rules and visualizes metrics, logs, traces in a beautiful web UI.
Group: monitoring
License: GPL

%package pantheonctl
Summary: The commond line interface of pantheon.
Group: pantheon/cli
Vendor: cylonchau
Source0: target/pantheonctl
URL: https://github.com/cylonchau/pantheon
BuildRoot: %{_tmppath}/%{name}-%{version}-buildroot

%description
Pantheon - Prometheus universal exporter cli

%description ctl
Pantheon cli and universal exporter cli


%define __arch_install_post %{nil}
%define __os_install_post %{nil}
%global debug_package %{nil}

%prep

%install
rm -rf %{buildroot}
%{__install} -p -D %{SOURCE0} %{buildroot}/usr/sbin/pantheonctl


%files pantheonctl
%defattr(-,root,root,-)
%attr(0555,root,n9e) /usr/sbin/pantheonctl

%changelog pantheonctl
* Thu Aug 22 2024 Cylon Chau <cylonchau@outlook.com>
- package pantheonctl