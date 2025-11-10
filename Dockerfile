FROM debian:stable-slim

# Install ldnsutils
RUN apt-get update && apt-get install -y ldnsutils && apt-get clean

# Set a working directory inside the container
WORKDIR /zones

# Default command: just open a shell (overridable)
CMD ["/bin/bash"]