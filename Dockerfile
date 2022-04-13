# Note that multi-stage builds should not be used in the ALM pipeline as GitLab
# has its own (and very cool) mechanism to transfer artifacts between jobs. See
# https://git.mgmt.innovo-cloud.de/help/ci/pipelines/job_artifacts

FROM ubuntu
WORKDIR /
COPY build/sim-exporter ./
ENTRYPOINT ["/sim-exporter"]