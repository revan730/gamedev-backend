FROM nginx
COPY nginx.conf /etc/nginx/nginx.conf
RUN ["mkdir", "/var/www"]
RUN ["mkdir", "/var/www/frontend"]
RUN ["chown", "nginx:nginx", "/var/www/frontend/"]
RUN ["ls", "-al", "/var/www"]