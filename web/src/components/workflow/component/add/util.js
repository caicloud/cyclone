const tramformArg = (data, isCustomStage) => {
  const value = _.cloneDeep(data);
  const containers = _.get(value, 'spec.containers[0]');
  _.forEach(containers, (v, k) => {
    if (['args', 'command'].includes(k) && _.isString(v)) {
      value.spec.containers[0][k] = v.split(/(?:\r\n|\r|\n)/);
      if (isCustomStage && k === 'command') {
        value.spec.containers[0][k] = _.concat(
          ['/bin/sh', '-e', '-c'],
          value.spec.containers[0][k]
        );
      }
    }
  });
  return value;
};

export const revertFormArg = (data, isCustom) => {
  const value = _.cloneDeep(data);
  const containers = _.get(value, 'spec.containers[0]');
  _.forEach(containers, (v, k) => {
    if (k === 'command' && _.isArray(v) && v.length > 0) {
      if (isCustom) {
        value.spec.containers[0][k] = _.drop(v, 3);
      }
      value.spec.containers[0][k] = value.spec.containers[0][k].join('\n');
    }
  });
  return value;
};

export const formatStage = data => {
  const isCustomStage = !_.get(data, [
    'metadata',
    'annotations',
    'stageTemplate',
  ]);
  let stage = {
    metadata: _.get(data, 'metadata'),
    spec: {
      pod: tramformArg(
        _.pick(_.get(data, 'spec.pod'), ['inputs', 'outputs', 'spec']),
        isCustomStage
      ),
    },
  };
  return stage;
};

export const formatSubmitData = (value, projectName, state) => {
  const { position, depend } = state;
  const requests = [];
  const stages = _.get(value, 'stages', []);
  const workflowInfo = {
    metadata: {
      name: _.get(value, 'metadata.name'),
      annotations: {
        description: _.get(value, 'metadata.annotations.description'),
        'cyclone.dev/owner': 'admin',
        stagePosition: JSON.stringify(position),
      },
    },
    spec: {
      stages: [],
    },
  };
  _.forEach(stages, v => {
    const currentStage = _.get(value, v);
    const stageFormatData = formatStage(currentStage);
    requests.push({
      type: 'createStage',
      project: projectName,
      data: stageFormatData,
    });

    const workflowStage = {
      artifacts: _.get(currentStage, 'spec.pod.outputs.artifacts', []),
      depends: _.get(depend, v),
      name: _.get(currentStage, `metadata.name`),
    };
    workflowInfo.spec.stages.push(workflowStage);
  });
  requests.push({
    type: 'createWorkflow',
    project: projectName,
    data: workflowInfo,
  });
  return requests;
};

export const revertWorkflow = data => {
  const workflow = {
    ..._.pick(data, ['metadata.name', 'metadata.annotations.description']),
    stages: [],
    currentStage: '',
  };
  _.forEach(_.get(data, 'spec.stages'), (t, k) => {
    workflow.stages.push(`stage_${k}`);
  });
  return workflow;
};
