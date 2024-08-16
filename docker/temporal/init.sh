#!/bin/sh

PATH="$PATH:/root/.temporalio/bin" >> ~/.bashrc

temporal operator namespace create "$TEMPORAL_NAMESPACE" || true

temporal operator search-attribute create --namespace "$TEMPORAL_NAMESPACE" \
		--name account_id --type Keyword \
