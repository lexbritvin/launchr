import * as core from "@actions/core";
import * as exec from "@actions/exec";
import * as fs from "fs";

async function cleanup() {
  try {
    // Get sidecar ID from state
    const sidecarId = core.getState('sidecar-id');

    if (!sidecarId) {
      core.warning('No sidecar ID found in state, nothing to clean up');
      return;
    }

    core.info(`Shutting down sidecar with ID: ${sidecarId}`);

    // Create shutdown signal file
    fs.writeFileSync('shutdown-signal.txt', 'shutdown=true');

    // Upload shutdown signal artifact
    const uploadParams = [
      'run', 'upload',
      '--name', `sidecar-${sidecarId}-shutdown`,
      '--repo', process.env.GITHUB_REPOSITORY,
      'shutdown-signal.txt'
    ];

    await exec.exec('gh', uploadParams);

    core.info(`Shutdown signal sent to sidecar ${sidecarId}`);

  } catch (error) {
    core.warning(`Cleanup failed with error: ${error.message}`);
    // We don't want to fail the workflow if cleanup fails
  }
}

await cleanup()