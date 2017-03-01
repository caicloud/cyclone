FROM 9.0-jre8-alpine

# change target/*.war to your path to war
COPY target/*.war $CATALINA_HOME/webapps/ROOT.war
