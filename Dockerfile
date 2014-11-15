FROM golang:1.3.3-onbuild
CMD ["./gondalf"]
CMD bash -C "./startApp.sh"
EXPOSE 3000