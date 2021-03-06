%define debug_package %{nil}

Summary:        Authenticated Metrics Proxy for Carbon/Graphite
Name:           amproxy
Version:        %{version}
Release:        1%{?dist}
License:        MIT
Group:          Development/Languages
URL:            https://github.com/jasonhancock/amproxy
Source0:        %{name}-%{version}.tar.gz
BuildRoot:      %{_tmppath}/%{name}-%{version}-%{release}-root-%(%{__id_u} -n)
Requires:       daemonize

%description
Authenticate Metrics Proxy for Carbon/Graphite.

%prep
%setup -q -n %{name}-%{version}

%build

export GOPATH=$RPM_BUILD_DIR/%{name}-%{version}
cd src/github.com/jasonhancock/amproxy && make


%install
rm -rf $RPM_BUILD_ROOT
mkdir -p $RPM_BUILD_ROOT/usr/sbin
install -m 0755 $RPM_BUILD_DIR/%{name}-%{version}/bin/amproxy $RPM_BUILD_ROOT/usr/sbin/

mkdir -p $RPM_BUILD_ROOT/%{_sysconfdir}/amproxy
mkdir -p $RPM_BUILD_ROOT/%{_sysconfdir}/rc.d/init.d
mkdir -p $RPM_BUILD_ROOT/%{_sysconfdir}/logrotate.d
mkdir -p $RPM_BUILD_ROOT/%{_sysconfdir}/sysconfig
install -m 0644 $RPM_BUILD_DIR/%{name}-%{version}/src/github.com/jasonhancock/amproxy/packaging/redhat/auth_file.yaml $RPM_BUILD_ROOT/%{_sysconfdir}/amproxy/auth_file.yaml
install -m 0755 $RPM_BUILD_DIR/%{name}-%{version}/src/github.com/jasonhancock/amproxy/packaging/redhat/amproxy.init $RPM_BUILD_ROOT/%{_sysconfdir}/rc.d/init.d/amproxy
install -m 0755 $RPM_BUILD_DIR/%{name}-%{version}/src/github.com/jasonhancock/amproxy/packaging/redhat/amproxy.logrotate $RPM_BUILD_ROOT/%{_sysconfdir}/logrotate.d/amproxy
install -m 0755 $RPM_BUILD_DIR/%{name}-%{version}/src/github.com/jasonhancock/amproxy/packaging/redhat/amproxy.sysconfig $RPM_BUILD_ROOT/%{_sysconfdir}/sysconfig/amproxy

mkdir -p $RPM_BUILD_ROOT%{_localstatedir}/lib/amproxy
mkdir -p $RPM_BUILD_ROOT%{_localstatedir}/log/amproxy

%pre
# Add the "amproxy" user
getent group amproxy >/dev/null || groupadd -r amproxy
getent passwd amproxy >/dev/null || \
  useradd -r -g amproxy -s /sbin/nologin \
    -d %{_localstatedir}/lib/amproxy -c "amproxy" amproxy
exit 0

%post
# Register the httpd service
/sbin/chkconfig --add amproxy

%preun
if [ $1 = 0 ]; then
    /sbin/service amproxy stop > /dev/null 2>&1
    /sbin/chkconfig --del amproxy
fi

%clean
rm -rf $RPM_BUILD_ROOT

%files
%defattr(-,root,root,-)
/usr/sbin/amproxy
%{_sysconfdir}/rc.d/init.d/amproxy
%config(noreplace) %{_sysconfdir}/amproxy/auth_file.yaml
%config(noreplace) %{_sysconfdir}/logrotate.d/amproxy
%config(noreplace) %{_sysconfdir}/sysconfig/amproxy

%attr(0700,amproxy,amproxy) %dir %{_localstatedir}/lib/amproxy
%attr(0700,amproxy,amproxy) %dir %{_localstatedir}/log/amproxy
