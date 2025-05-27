import * as core from "@actions/core";
import * as exec from "@actions/exec";
import * as path from "path";
import * as fs from "fs";
import * as crypto from "crypto";

async function run() {
  try {
    // Generate a unique ID for this sidecar instance
    const sidecarId = crypto.randomUUID();
    core.setOutput('sidecar-id', sidecarId);

    // Save sidecar ID for post cleanup
    core.saveState('sidecar-id', sidecarId);

    // Get timeout from inputs
    const timeout = core.getInput('timeout');

    core.info(`Starting Linux Docker sidecar with ID: ${sidecarId}`);

    // Check for GitHub CLI
    try {
      await exec.exec('gh', ['--version']);
    } catch (error) {
      core.error('GitHub CLI not found. Please install it using actions/setup-gh');
      throw new Error('GitHub CLI is required for this action');
    }

    // Start the Linux sidecar workflow
    const workflowParams = [
      'workflow', 'run', 'linux-sidecar.yml',
      '-f', `sidecar_id=${sidecarId}`,
      '-f', `timeout=${timeout}`,
      '--ref', process.env.GITHUB_REF,
      '--repo', process.env.GITHUB_REPOSITORY
    ];

    await exec.exec('gh', workflowParams);

    // Wait for sidecar to start and provide connection details
    core.info('Waiting for sidecar to start...');

    const maxAttempts = 30;
    let attempt = 0;
    let sidecarStarted = false;

    while (attempt < maxAttempts) {
      attempt++;

      try {
        // Try to download the connection details artifact
        const downloadParams = [
          'run', 'download',
          '--name', `sidecar-${sidecarId}-details`,
          '--repo', process.env.GITHUB_REPOSITORY
        ];

        const exitCode = await exec.exec('gh', downloadParams, { ignoreReturnCode: true });

        if (exitCode === 0) {
          core.info('Sidecar started successfully');
          sidecarStarted = true;
          break;
        }
      } catch (error) {
        // Ignore error and retry
      }

      core.info(`Waiting for sidecar to start (attempt ${attempt}/${maxAttempts})...`);
      await new Promise(resolve => setTimeout(resolve, 10000)); // Wait 10 seconds
    }

    if (!sidecarStarted) {
      throw new Error('Timed out waiting for sidecar to start');
    }

    // Read connection details
    const details = fs.readFileSync('sidecar-details.env', 'utf8');
    const dockerHost = details.match(/DOCKER_HOST=(.*)/)[1];

    core.setOutput('docker-host', dockerHost);
    core.exportVariable('DOCKER_HOST', dockerHost);

    // Download Docker certificates
    const certParams = [
      'run', 'download',
      '--name', 'docker-certs',
      '--repo', process.env.GITHUB_REPOSITORY
    ];

    await exec.exec('gh', certParams);

    // Set certificate path
    const certPath = path.join(process.cwd(), 'docker-certs');
    core.setOutput('docker-cert-path', certPath);
    core.exportVariable('DOCKER_CERT_PATH', certPath);

    // Set TLS verification
    if (core.getInput('tlsVerify') === 'true') {
      core.exportVariable('DOCKER_TLS_VERIFY', '1');
    }

    // Test Docker connection
    core.info('Testing Docker connection...');
    await exec.exec('docker', ['info']);

    core.info('Docker sidecar setup complete');

  } catch (error) {
    core.setFailed(`Action failed with error: ${error.message}`);
  }
}

await run()
