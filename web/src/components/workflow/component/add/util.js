import { tranformStage } from '@/lib/util';

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

export const formatSubmitData = (value, query, state) => {
  const { position, depend } = state;
  const requests = [];
  const stages = _.get(value, 'stages', []);
  const workflowInfo = {
    metadata: {
      name: _.get(value, 'metadata.name'),
      annotations: {
        stagePosition: JSON.stringify(position),
      },
    },
    spec: {
      stages: [],
    },
  };
  _.forEach(stages, v => {
    const inputResources = _.get(value, `${v}.inputs.resources`, []);
    const outputResources = _.get(value, `${v}.outputs.resources`, []);
    let stage = {
      metadata: _.get(value, 'metadata'),
      spec: {
        pod: tramformArg(
          _.pick(_.get(value, v), ['inputs', 'outputs', 'spec'])
        ),
      },
    };
    _.forEach(inputResources, (r, i) => {
      const data = _.pick(r, ['spec', 'metadata.name']);
      stage.spec.pod.inputs.resources[i] = _.pick(r, ['name', 'path']);
      requests.push({ type: 'createResource', project: query.project, data });
    });

    _.forEach(outputResources, (r, i) => {
      const data = _.pick(r, ['spec', 'metadata.name']);
      stage.spec.pod.outputs.resources[i] = _.pick(r, ['name']);
      requests.push({ type: 'createResource', project: query.project, data });
    });
    requests.push({
      type: 'createStage',
      project: query.project,
      data: stage,
    });

    const workflowStage = {
      artifacts: _.get(v, 'outputs.artifacts', []),
      depends: _.get(depend, v),
      name: _.get(value, `${v}.name`),
    };
    workflowInfo.spec.stages.push(workflowStage);
  });
  requests.push({
    type: 'createWorkflow',
    project: query.project,
    data: workflowInfo,
  });
  console.log('******', JSON.stringify(workflowInfo));
  return requests;
};

export const revertWorkflow = data => {
  const workflow = {
    metadata: {
      name: _.get(data, 'metadata.name'),
      description: _.get(data, 'metadata.annotations.description'),
    },
    stages: [],
    currentStage: '',
  };
  _.forEach(_.get(data, 'spec.stages'), (t, k) => {
    workflow.stages.push(`stage_${k}`);
  });
  return workflow;
};
