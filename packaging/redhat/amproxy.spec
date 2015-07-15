Summary:        Authenticated Metrics Proxy for Carbon/Graphite
Name:           amproxy
Version:        %{version}
Release:        1%{?dist}
License:        MIT
Group:          Development/Languages
URL:            https://github.com/jasonhancock/amproxy
Source:         %{name}-%{version}.tar.gz
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


%clean
rm -rf $RPM_BUILD_ROOT

%files
%defattr(-,root,root,-)
/usr/bin/amproxy
