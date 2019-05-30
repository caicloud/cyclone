import * as React from 'react';
import { Button, Drawer, notification } from 'antd';
import { GraphView } from 'react-digraph';
import GraphConfig, {
  NODE_KEY,
  STAGE,
  EMPTY_EDGE_TYPE,
  SPECIAL_EDGE_TYPE,
} from './graph-config'; // Configures node/edge types
import AddStage from '../stage/AddStage';
import classNames from 'classnames';
import styles from '../index.module.less';
import PropTypes from 'prop-types';
import { formatStage } from './util';
import { formatTouchedField } from '@/lib/util';
import { inject, observer } from 'mobx-react';

// NOTE: Edges must have 'source' & 'target' attributes
// In a more realistic use case, the graph would probably originate
// elsewhere in the App or be generated from some other state upstream of this component.
@inject('resource', 'workflow')
@observer
class Graph extends React.Component {
  GraphView;

  constructor(props) {
    super(props);

    const { initialGraph } = props;

    this.state = {
      copiedNode: null,
      graph: _.isEmpty(initialGraph)
        ? {
            edges: [],
            nodes: [],
          }
        : initialGraph,
      nodePosition: this.getNodePosition(initialGraph),
      selected: null,
      visible: false,
      stageInfo: {},
      depnedLoading: false,
    };

    this.GraphView = React.createRef();
  }

  componentWillUnmount() {
    const { saveGraphWhenUnmount } = this.props;
    const { graph } = this.state;
    saveGraphWhenUnmount(graph);
  }

  getNodePosition = initialGraph => {
    if (!_.isEmpty(initialGraph)) {
      const nodes = _.get(initialGraph, 'nodes');
      const pos = {};
      _.forEach(nodes, v => {
        pos[v.id] = _.pick(v, ['x', 'y', 'title']);
      });
      return pos;
    }
    return {};
  };

  // Helper to find the index of a given node
  getNodeIndex(searchNode) {
    return this.state.graph.nodes.findIndex(node => {
      return node[NODE_KEY] === searchNode[NODE_KEY];
    });
  }

  // Helper to find the index of a given edge
  getEdgeIndex(searchEdge) {
    return this.state.graph.edges.findIndex(edge => {
      return (
        edge.source === searchEdge.source && edge.target === searchEdge.target
      );
    });
  }

  // Given a nodeKey, return the corresponding node
  getViewNode(nodeKey) {
    const searchNode = {};
    searchNode[NODE_KEY] = nodeKey;
    const i = this.getNodeIndex(searchNode);
    return this.state.graph.nodes[i];
  }

  addOrUpdateStageOnClose = () => {
    const { graph, selected, nodePosition, stageInfo } = this.state;
    const {
      values,
      setFieldValue,
      updateStagePosition,
      update,
      project,
      resource: { updateStage, createStage },
    } = this.props;
    let _state = {
      graph,
      visible: false,
      nodePosition,
    };

    const stages = _.get(values, 'stages', []);
    const currentStage = _.get(values, 'currentStage', '');
    const number = currentStage.split('_')[1];
    const stageName = _.get(values, `${currentStage}.metadata.name`);
    // using a new array like this creates a new memory reference
    // this will force a re-render
    if (!selected && !stages.includes(currentStage)) {
      const position = {
        title: stageName,
        x: 100 + number * 140, // 动态随机定位
        y: 100 + number * 60,
      };
      _state.graph.nodes = [
        {
          id: currentStage, // NOTE: stage id
          type: STAGE,
          ...position,
        },
        ...graph.nodes,
      ];
      _state.nodePosition[currentStage] = position;
      setFieldValue('stages', [...stages, currentStage]);
      updateStagePosition(currentStage, {
        ...position,
      });
      // create stage
      if (update) {
        const stage = formatStage(
          _.get(values, `${_.get(values, 'currentStage')}`),
          false
        );
        createStage(project, stage, data => {
          // update workflow stages after add stage
          const workflowStage = {
            name: _.get(data, 'metadata.name'),
            artifacts: _.get(data, 'spec.pod.outputs.artifacts', []),
            depends: [],
          };
          this.updateStageDepned(workflowStage, '', 'addStage');
        });
      }
    }

    if (selected && stages.includes(currentStage)) {
      const nodeTitle = _.get(values, `${currentStage}.metadta.name`);
      const stageNode = this.getViewNode(currentStage);
      if (nodeTitle && nodeTitle !== _.get(stageNode, 'title')) {
        const i = this.getNodeIndex(stageNode);
        _state.graph.nodes[i].title = nodeTitle;
        // TODO(qme): update stage depend
      }

      if (update) {
        const stage = formatStage(_.get(values, currentStage), false);
        const prevStageInfo = _.get(stageInfo, stageName);
        if (!_.isEqual(stage, prevStageInfo)) {
          updateStage(project, stageName, stage, () => {
            notification.success({
              message: intl.get('notification.updateStage'),
              duration: 2,
            });
          });
        }
      }
    }
    this.setState(_state);
  };

  onClose = () => {
    const { values, validateForm, setTouched } = this.props;
    validateForm(_.get(values, _.get(values, 'currentStage'), {})).then(
      error => {
        if (_.isEmpty(error)) {
          this.addOrUpdateStageOnClose();
        } else {
          const fieldObj = formatTouchedField(error);
          setTouched(fieldObj);
        }
      }
    );
  };

  getStageId = array => {
    const number = array.map(o => o.split('_')[1]);
    const max = number.sort(function(a, b) {
      return b - a;
    })[0];
    return max * 1 + 1 || 0;
  };

  addStartNode = () => {
    const { setFieldValue, values } = this.props;
    // show Drawer
    this.setState({ visible: true, selected: null });
    // TODO(qme): stage id random
    const stageId = `stage_${this.getStageId(_.get(values, 'stages'))}`;
    setFieldValue('currentStage', stageId);
    setFieldValue(stageId, {
      metadata: { name: '' },
      spec: {
        pod: {
          inputs: {
            resources: [],
          },
          spec: {
            containers: [
              {
                args: [],
                command: [],
                image: '',
                env: [],
              },
            ],
          },
        },
      },
    });
  };

  /*
   * Handlers/Interaction
   */

  // Called by 'drag' handler, etc..
  // to sync updates from D3 with the graph
  onUpdateNode = viewNode => {
    const { updateStagePosition } = this.props;
    const graph = this.state.graph;
    const i = this.getNodeIndex(viewNode);
    graph.nodes[i] = viewNode;
    updateStagePosition(viewNode.id, _.pick(viewNode, ['title', 'x', 'y']));
    this.setState({ graph });
  };

  // Node 'mouseUp' handler
  onSelectNode = viewNode => {
    const { nodePosition, stageInfo, depnedLoading } = this.state;
    const {
      setFieldValue,
      update,
      resource: { getStage },
      project,
      values,
    } = this.props;
    if (depnedLoading) {
      return;
    }
    // Deselect events will send Null viewNode
    let state = { selected: viewNode, nodePosition, stageInfo };
    const nodeId = _.get(viewNode, 'id');
    const moved =
      _.get(nodePosition, `${nodeId}.x`) !== _.get(viewNode, 'x') ||
      _.get(nodePosition, `${nodeId}.y`) !== _.get(viewNode, 'y');
    if (viewNode && !moved) {
      setFieldValue('currentStage', nodeId);
      if (update) {
        const stageName = _.get(viewNode, 'title');
        getStage(project, stageName, data => {
          const info = _.pick(data, [
            'metadata.name',
            'metadata.annotations.stageTemplate',
            'spec',
          ]);
          if (!_.get(values, nodeId)) {
            setFieldValue(nodeId, info);
          }
          this.setState({ stageInfo: { [stageName]: info }, visible: true });
        });
      } else {
        state.visible = true;
      }
    } else {
      state.nodePosition[nodeId] = _.pick(viewNode, ['x', 'y', 'title']);
    }

    this.setState(state);
  };

  // Edge 'mouseUp' handler
  onSelectEdge = viewEdge => {
    this.setState({ selected: viewEdge });
  };

  // Updates the graph with a new node
  onCreateNode = (x, y) => {
    const graph = this.state.graph;

    const viewNode = {
      id: Date.now(),
      title: '',
      type: STAGE,
      x,
      y,
    };
    graph.nodes = [...graph.nodes, viewNode];
    this.setState({ graph });
  };

  // Deletes a node from the graph
  onDeleteNode = (viewNode, nodeId, nodeArr) => {
    const {
      update,
      updateStagePosition,
      setFieldValue,
      values,
      setStageDepned,
    } = this.props;
    const { graph, nodePosition } = this.state;
    // Delete any connected edges
    const newEdges = graph.edges.filter((edge, i) => {
      return (
        edge.source !== viewNode[NODE_KEY] && edge.target !== viewNode[NODE_KEY]
      );
    });
    const stages = _.get(values, 'stages', []);
    const index = stages.indexOf(nodeId);
    if (index > -1) {
      _.pullAt(stages, index);
      setFieldValue('stages', stages);
      setFieldValue(nodeId, {});
    }
    setStageDepned(nodeId, '', _.get(viewNode, 'title'));
    const position = _.cloneDeep(nodePosition);
    graph.nodes = nodeArr;
    graph.edges = newEdges;
    delete position[nodeId];
    this.setState({ graph, selected: null, nodePosition: position });
    updateStagePosition(nodeId, {});
    if (update) {
      this.updateStageDepned(_.get(viewNode, 'title'), '', 'removeNode');
    }
  };

  judgeEdgeCricle = (source, target, edge) => {
    let circle = false;
    const checkCircle = _target => {
      let sources = [];
      _.forEach(edge, v => {
        if (v.source === _target) {
          sources.push(v);
        }
      });
      _.forEach(sources, v => {
        if (v.target === source) {
          circle = true;
        }
        if (!circle) {
          checkCircle(v.target);
        }
      });
    };
    checkCircle(target);
    return circle;
  };

  // Creates a new node between two edges
  onCreateEdge = (sourceViewNode, targetViewNode) => {
    const { setStageDepned, update } = this.props;
    const graph = this.state.graph;
    // This is just an example - any sort of logic
    // could be used here to determine edge type
    const type =
      sourceViewNode.type === STAGE ? SPECIAL_EDGE_TYPE : EMPTY_EDGE_TYPE;

    const viewEdge = {
      source: sourceViewNode[NODE_KEY],
      target: targetViewNode[NODE_KEY],
      type,
    };

    // Determine whether a closed loop is formed
    const isCircle = this.judgeEdgeCricle(
      sourceViewNode[NODE_KEY],
      targetViewNode[NODE_KEY],
      graph.edges
    );

    // Only add the edge when the source node is not the same as the target
    if (viewEdge.source !== viewEdge.target && !isCircle) {
      graph.edges = [...graph.edges, viewEdge];
      this.setState({
        graph,
        selected: viewEdge,
      });
      setStageDepned(
        _.get(targetViewNode, 'id'),
        _.get(sourceViewNode, 'title')
      );
      if (update) {
        this.updateStageDepned(
          _.get(targetViewNode, 'title'),
          _.get(sourceViewNode, 'title'),
          'addEdge'
        );
      }
    }
  };

  updateStageDepned = (targetName, sourceName, type) => {
    const {
      workflowName,
      values,
      project,
      workflow: { updateWorkflow, workflowDetail },
    } = this.props;
    const { nodePosition } = this.state;
    this.setState({ depnedLoading: true });
    const detail = _.get(workflowDetail, workflowName);
    const workflowInfo = _.cloneDeep({
      ..._.pick(values, ['metadata']),
      ..._.pick(detail, 'spec.stages'),
    });
    workflowInfo.metadata.annotations.stagePosition = JSON.stringify(
      nodePosition
    );
    const index = _.findIndex(
      workflowInfo.spec.stages,
      s => s.name === targetName
    );
    if (index > -1 || type === 'addStage') {
      const depends =
        _.get(workflowInfo, 'spec.spec.stages[index].depends') || [];
      switch (type) {
        case 'addEdge': {
          workflowInfo.spec.stages[index].depends = depends;
          workflowInfo.spec.stages[index].depends.push(sourceName);
          break;
        }
        case 'removeEdge': {
          const removedIndex = workflowInfo.spec.stages[index].depends.indexOf(
            sourceName
          );
          _.pullAt(workflowInfo.spec.stages[index].depends, removedIndex);
          break;
        }
        case 'removeNode': {
          _.pullAt(workflowInfo.spec.stages, index);
          _.forEach(workflowInfo.spec.stages, s => {
            if (s.depends && s.depends.includes(targetName)) {
              const i = s.depends.indexOf(targetName);
              _.pullAt(s.depends, i);
            }
          });
          break;
        }
        case 'addStage': {
          workflowInfo.spec.stages.push(targetName);
          break;
        }
        default: {
          break;
        }
      }
    }
    updateWorkflow(project, workflowName, workflowInfo, () => {
      // NOTE: 停留时间太短，看不清
      notification.success({
        message: intl.get('notification.updateWorkflow'),
        duration: 2,
      });
      this.setState({ depnedLoading: false });
    });
  };

  // Called when an edge is reattached to a different target.
  onSwapEdge = (sourceViewNode, targetViewNode, viewEdge) => {
    const graph = this.state.graph;
    const i = this.getEdgeIndex(viewEdge);
    const edge = JSON.parse(JSON.stringify(graph.edges[i]));

    edge.source = sourceViewNode[NODE_KEY];
    edge.target = targetViewNode[NODE_KEY];
    graph.edges[i] = edge;
    // reassign the array reference if you want the graph to re-render a swapped edge
    graph.edges = [...graph.edges];
    this.setState({
      graph,
      selected: edge,
    });
  };

  // Called when an edge is deleted
  onDeleteEdge = (viewEdge, edges) => {
    const { setStageDepned, update } = this.props;
    const {
      graph: { nodes },
    } = this.state;
    const sourceNode = _.find(
      nodes,
      n => _.get(n, 'id') === _.get(viewEdge, 'source')
    );
    setStageDepned(_.get(viewEdge, 'target'), _.get(sourceNode, 'title'));
    const graph = this.state.graph;
    graph.edges = edges;
    this.setState({
      graph,
      selected: null,
    });

    if (update) {
      const targetName = _.get(
        _.find(nodes, n => _.get(n, 'id') === _.get(viewEdge, 'target')),
        'title'
      );
      const sourceName = _.get(
        _.find(nodes, n => _.get(n, 'id') === _.get(viewEdge, 'source')),
        'title'
      );
      this.updateStageDepned(targetName, sourceName, 'removeEdge');
    }
  };

  onCopySelected = () => {
    if (this.state.selected.source) {
      console.warn('Cannot copy selected edges, try selecting a node instead.'); //eslint-disable-line
      return;
    }
    const x = this.state.selected.x + 10;
    const y = this.state.selected.y + 10;
    this.setState({
      copiedNode: { ...this.state.selected, x, y },
    });
  };

  onPasteSelected = () => {
    if (!this.state.copiedNode) {
      console.warn('No node is current in the copy queue.'); //eslint-disable-line
    }
    const graph = this.state.graph;
    const newNode = { ...this.state.copiedNode, id: Date.now() };
    graph.nodes = [...graph.nodes, newNode];
    this.forceUpdate();
  };

  // render note text
  renderNodeText = (data, id, isSelected) => {
    const lineOffset = 6;
    const cls = classNames('node-text', { selected: isSelected });
    return (
      <text className={cls} textAnchor="middle">
        {!!data.typeText && <tspan opacity="0.5">{data.typeText}</tspan>}
        {data.title && (
          <tspan x={0} dy={lineOffset} fontSize="16px">
            {data.title}
          </tspan>
        )}
      </text>
    );
  };

  /*
   * Render
   */

  render() {
    const {
      graph: { nodes, edges },
      stageInfo,
      depnedLoading,
      selected,
    } = this.state;
    const { NodeTypes, NodeSubtypes, EdgeTypes } = GraphConfig;
    const { values, update, project } = this.props;
    const currentStage = _.get(values, 'currentStage');
    const stages = _.get(values, 'stages', []);
    const modify = stages.includes(currentStage);
    const stageName = update
      ? _.get(
          stageInfo,
          `${_.get(selected, 'title')}.metadata.annotations.stageTemplate`
        )
      : '';
    return (
      <div id="graph" className={styles['graph']}>
        <div className="graph-header">
          <Button
            type="primary"
            onClick={this.addStartNode}
            disabled={depnedLoading}
          >
            添加 stage
          </Button>
        </div>
        <GraphView
          ref={el => (this.GraphView = el)}
          nodeKey={NODE_KEY}
          nodes={nodes}
          edges={edges}
          selected={selected}
          nodeTypes={NodeTypes}
          nodeSubtypes={NodeSubtypes}
          edgeTypes={EdgeTypes}
          onSelectNode={this.onSelectNode}
          onCreateNode={this.onCreateNode}
          onUpdateNode={this.onUpdateNode}
          onDeleteNode={this.onDeleteNode}
          onSelectEdge={this.onSelectEdge}
          onCreateEdge={this.onCreateEdge}
          onSwapEdge={this.onSwapEdge}
          onDeleteEdge={this.onDeleteEdge}
          onCopySelected={this.onCopySelected}
          onPasteSelected={this.onPasteSelected}
          renderNodeText={this.renderNodeText}
          readOnly={depnedLoading}
          showGraphControls={false}
        />
        <Drawer
          title={
            modify
              ? `${intl.get('operation.modify')} stage`
              : `${intl.get('operation.add')} stage`
          }
          placement="right"
          closable={false}
          onClose={this.onClose}
          visible={this.state.visible}
          width={800}
        >
          <AddStage
            key={_.get(values, 'currentStage')}
            setFieldValue={this.props.setFieldValue}
            values={this.props.values}
            update={update}
            project={project}
            templateName={stageName}
          />
        </Drawer>
      </div>
    );
  }
}

Graph.propTypes = {
  values: PropTypes.object,
  setFieldValue: PropTypes.func,
  setStageDepned: PropTypes.func,
  updateStagePosition: PropTypes.func,
  initialGraph: PropTypes.object,
  update: PropTypes.bool,
  project: PropTypes.string,
  workflowName: PropTypes.string,
  resource: PropTypes.object,
  saveGraphWhenUnmount: PropTypes.func,
  workflow: PropTypes.object,
  validateForm: PropTypes.func,
  setTouched: PropTypes.func,
};

export default Graph;
