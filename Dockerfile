FROM gcr.io/distroless/static-debian11:nonroot
ENTRYPOINT ["/baton-successfactors"]
COPY baton-successfactors /