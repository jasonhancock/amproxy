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

%description
Authenticate Metrics Proxy for Carbon/Graphite.

%prep
%setup -q -n %{name}-%{version}

%build

export GOPATH=$RPM_BUILD_DIR/%{name}-%{version}
make

%install
rm -rf $RPM_BUILD_ROOT
mkdir -p $RPM_BUILD_ROOT/usr/bin
install -m 0755 $RPM_BUILD_DIR/%{name}-%{version}/bin/amproxy $RPM_BUILD_ROOT/usr/bin/

mkdir -p $RPM_BUILD_ROOT/etc/rc.d/init.d
install -m 0755 $RPM_BUILD_DIR/%{name}-%{version}/packaging/redhat/amproxy.init $RPM_BUILD_ROOT/etc/rc.d/init.d/amproxy

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
/usr/bin/amproxy
%dir %{_localstatedir}/log/amproxy
%{_sysconfdir}/rc.d/init.d/amproxy

%attr(0700,amproxy,amproxy) %dir %{_localstatedir}/lib/amproxy
