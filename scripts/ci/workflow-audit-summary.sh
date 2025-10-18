#!/usr/bin/env bash
set -euo pipefail

# Append a short summary of the workflow audit to GitHub Step Summary.
{
  echo "## Workflow Audit Summary"
  echo "Checked .github/workflows for master/main/latest pins and curl|sh installers."
} >> "${GITHUB_STEP_SUMMARY}"
