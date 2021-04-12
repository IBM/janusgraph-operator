FROM cassandra:3.11.8
  
RUN chgrp -R root /etc/cassandra
RUN chmod -R g+w /etc/cassandra

RUN grep -v "rpc_address:" /etc/cassandra/cassandra.yaml > tmpfile && mv tmpfile /etc/cassandra/cassandra.yaml
RUN grep -v "broadcast_rpc_address:" /etc/cassandra/cassandra.yaml > tmpfile && mv tmpfile /etc/cassandra/cassandra.yaml
RUN grep -v "authenticator:" /etc/cassandra/cassandra.yaml > tmpfile && mv tmpfile /etc/cassandra/cassandra.yaml

RUN echo "rpc_address: 0.0.0.0" >> /etc/cassandra/cassandra.yaml
RUN echo "broadcast_rpc_address: 1.2.3.4" >> /etc/cassandra/cassandra.yaml
RUN echo "authenticator: PasswordAuthenticator" >> /etc/cassandra/cassandra.yaml

RUN sed  -i '1i LOCAL_JMX=no' /opt/cassandra/conf/cassandra-env.sh