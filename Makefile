VERSION=1.1.139

.PHONY: default
default: compile

OBJECTS=dhound-agent

.PHONY: compile
compile: $(OBJECTS)

dhound-agent:
	go build --ldflags '-extldflags "-static"' -o $@

.PHONY: clean
clean: 
	-rm $(OBJECTS)
	-rm -rf build
	-rm -rf publish/empty

.PHONY: buildempty
buildempty:
	-mkdir publish/empty


.PHONY: rpm deb
deb: AFTER_INSTALL=publish/pkg/ubuntu/after-install.sh
rpm: AFTER_INSTALL=publish/pkg/centos/after-install.sh
rpm: BEFORE_INSTALL=publish/pkg/centos/before-install.sh
rpm: BEFORE_REMOVE=publish/pkg/centos/before-remove.sh
deb: AFTER_INSTALL=publish/pkg/ubuntu/after-install.sh
deb: BEFORE_INSTALL=publish/pkg/ubuntu/before-install.sh
deb: BEFORE_REMOVE=publish/pkg/ubuntu/before-remove.sh
rpm deb: PREFIX=/opt/dhound-agent
rpm deb: clean compile buildempty
	fpm -f -s dir -t $@ -n dhound-agent -v $(VERSION) \
		--architecture $(ARCHITECTURE) \
		--replaces dhound-agent \
		--description "dhound-agent tool for collecting security events in the system" \
		--after-install $(AFTER_INSTALL) \
		--before-install $(BEFORE_INSTALL) \
		--before-remove $(BEFORE_REMOVE) \
		--config-files /etc/dhound-agent/config.yml \
		./dhound-agent=$(PREFIX)/bin/ \
		./config/config.sample.yml=/etc/dhound-agent/config.yml \
		./publish/etc=/ \
		./publish/empty/=/var/lib/dhound-agent/ \
		./publish/empty/=/var/log/dhound-agent/ \
		./publish/empty/=/etc/dhound-agent/rules.d/ \
		./config/rules.d/custom.yml=/etc/dhound-agent/rules.d/custom.yml \
		./config/rules.d/fail2ban.yml=/etc/dhound-agent/rules.d/fail2ban.yml \
		./config/rules.d/pure-ftpd.yml=/etc/dhound-agent/rules.d/pure-ftpd.yml \
		./config/rules.d/sshd.yml=/etc/dhound-agent/rules.d/sshd.yml \
		./config/rules.d/tcp-out.yml=/etc/dhound-agent/rules.d/tcp-out.yml \
		./config/rules.d/couchbase.yml=/etc/dhound-agent/rules.d/couchbase.yml \
		./config/rules.d/installations.yml=/etc/dhound-agent/rules.d/installations.yml \
		./config/rules.d/openvpn.yml=/etc/dhound-agent/rules.d/openvpn.yml \
		./config/rules.d/wordpress-accesslog.yml=/etc/dhound-agent/rules.d/wordpress-accesslog.yml \

