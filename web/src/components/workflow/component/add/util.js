const tramformArg = data => {
  const value = _.cloneDeep(data);
  const containers = _.get(value, 'spec.containers[0]');
  _.forEach(containers, (v, k) => {
    if (['args', 'command'].includes(k) && _.isString(v)) {
      value.spec.containers[0][k] = v.split(/(?:\r\n|\r|\n)/);
    }
  });
  return value;
};

export const formatStage = (data, fromCreate = true, requests, query) => {
  const inputResources = _.get(data, `spec.pod.inputs.resources`, []);
  const outputResources = _.get(data, `spec.pod.outputs.resources`, []);
  let stage = {
    metadata: _.get(data, 'metadata'),
    spec: {
      pod: tramformArg(
        _.pick(_.get(data, 'spec.pod'), ['inputs', 'outputs', 'spec'])
      ),
    },
  };
  _.forEach(inputResources, (r, i) => {
    stage.spec.pod.inputs.resources[i] = {
      ..._.pick(r, ['name', 'path', 'type']),
    };
    if (fromCreate) {
      const resourceData = {
        metadata: { name: _.get(r, 'name') },
        spec: {
          type: _.get(r, 'type'),
          ..._.get(r, 'spec'),
        },
      };
      requests.push({
        type: 'createResource',
        project: query.project,
        data: resourceData,
      });
    }
  });

  _.forEach(outputResources, (r, i) => {
    stage.spec.pod.outputs.resources[i] = {
      ..._.pick(r, ['name', 'type']),
      path: '',
    };
    if (fromCreate) {
      const resourceData = {
        metadata: { name: _.get(r, 'name') },
        spec: {
          type: _.get(r, 'type'),
          ..._.get(r, 'spec'),
        },
      };
      requests.push({
        type: 'createResource',
        project: query.project,
        data: resourceData,
      });
    }
  });

  return stage;
};

export const formatSubmitData = (value, query, state) => {
  const { position, depend } = state;
  const requests = [];
  const stages = _.get(value, 'stages', []);
  const workflowInfo = {
    metadata: {
      name: _.get(value, 'metadata.name'),
      annotations: {
        description: _.get(value, 'metadata.annotations.description'),
        stagePosition: JSON.stringify(position),
      },
    },
    spec: {
      stages: [],
    },
  };
  _.forEach(stages, v => {
    const currentStage = _.get(value, v);
    const stageFormatData = formatStage(currentStage, true, requests, query);
    requests.push({
      type: 'createStage',
      project: query.project,
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
    project: query.project,
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
