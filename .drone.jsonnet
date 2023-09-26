local image_name = 'nerfd/drone-email';

local ALPINE_VERSION = '3.18.3';
local BUILD_IMAGE_TAG = '1.21.1-alpine';

local docker_defaults = {
  username: {
    from_secret: 'docker_username',
  },
  password: {
    from_secret: 'docker_password',
  },
};

local pipeline_defaults = {
  kind: 'pipeline',
  type: 'docker',
};

local trigger = {
  trigger: {
    branch: [
      'master',
    ],
    event: [
      'push',
      'custom',
    ],
  },
};

local platform(arch) = {
  platform: {
    os: 'linux',
    arch: arch,
  },
};

local build_image(arch) = {
  name: 'build_image_' + arch,
  local GOARCH = [if arch == 'arm64' then 'arm' else 'amd64'],
  steps: [
    {
      name: 'build_image',
      image: 'plugins/docker',
      settings: {
        dockerfile: 'Dockerfile',
        label_schema: [
          'docker.dockerfile=Dockerfile',
        ],
        build_args: [
          'ALPINE_VERSION=' + ALPINE_VERSION,
          'BUILD_IMAGE_TAG=' + BUILD_IMAGE_TAG,
          'GOARCH=' + GOARCH,
        ],
        repo: image_name,
        tags: [
          'latest-' + arch,
        ],
      } + docker_defaults,
    },
  ],
} + platform(arch) + trigger + pipeline_defaults;

local signature_placeholder = {
  kind: 'signature',
  hmac: '0000000000000000000000000000000000000000000000000000000000000000',
};

local multiarch_image = {
  local image_tag = 'latest',
  name: 'multiarch_image',
  depends_on: [
    'build_image_amd64',
    'build_image_arm64',
  ],
  steps: [
    {
      name: 'multiarch_image',
      image: 'plugins/manifest',
      settings: {
        target: std.format('%s:%s', [image_name, image_tag]),
        template: std.format('%s:%s-ARCH', [image_name, image_tag]),
        platforms: [
          'linux/amd64',
          'linux/arm64',
        ],
      } + docker_defaults,
    },
  ],
} + trigger + pipeline_defaults;

[
  build_image('amd64'),
  build_image('arm64'),
  multiarch_image,
  signature_placeholder,
]
