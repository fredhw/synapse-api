FROM trestletech/plumber

RUN apt-get install zip unzip
RUN R -e "install.packages('e1071')"
RUN R -e "install.packages('pracma')"
RUN R -e "install.packages('jsonlite')"
RUN R -e "install.packages('base64enc')"
RUN R -e "install.packages('openxlsx')"

COPY . /app/

EXPOSE 80

ENTRYPOINT ["R", "-e", "setwd('/app/'); pr <- plumber::plumb(commandArgs()[4]); pr$run(host='0.0.0.0', port=80)"]
CMD ["/app/api.R"]
