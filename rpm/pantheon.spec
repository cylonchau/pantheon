Name: pantheon
Version: %{_version}
Release: 1%{?dist}
Summary: Pantheon - Prometheus target manager with HTTP SD and universal exporter cli.
Group: monitoring
License: Apache License 2.0
URL: https://github.com/cylonchau/pantheon
Source0: target/pantheonctl
BuildRoot: %{_tmppath}/%{name}-%{version}-buildroot

%description
Pantheon - Prometheus universal exporter cli

%package pantheonctl
Summary: The commond line interface of pantheon.
Group: pantheon/cli
Vendor: cylonchau

%description pantheonctl
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
%attr(0555,root,root) /usr/sbin/pantheonctl

%changelog pantheonctl
* Thu Aug 22 2024 Cylon Chau <cylonchau@outlook.com>
- package pantheonctl