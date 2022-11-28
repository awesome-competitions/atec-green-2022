FROM centos:7
RUN mkdir -p /home/admin/atec_project
WORKDIR /home/admin/atec_project
COPY app /home/admin/atec_project/app
COPY run.sh /home/admin/atec_project/run.sh