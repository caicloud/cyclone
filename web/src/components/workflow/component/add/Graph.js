import * as React from 'react';
import { Button, Drawer } from 'antd';
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

// NOTE: Edges must have 'source' & 'target' attributes
// In a more realistic use case, the graph would probably originate
// elsewhere in the App or be generated from some other state upstream of this component.
class Graph extends React.Component {
  GraphView;

  constructor(props) {
    super(props);

    this.state = {
      copiedNode: null,
      graph: {
        edges: [],
        nodes: [],
      },
      nodePosition: {},
      selected: null,
      visible: false,
    };

    this.GraphView = React.createRef();
  }

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

  onClose = () => {
    const { graph, selected, nodePosition } = this.state;
    const { values, setFieldValue } = this.props;

    let _state = {
      graph,
      visible: false,
      nodePosition,
    };

    const stages = _.get(values, 'stages', []);
    const currentStage = _.get(values, 'currentStage', '');
    const number = currentStage.split('_')[1] - 1;
    // using a new array like this creates a new memory reference
    // this will force a re-render
    if (!selected && !stages.includes(currentStage)) {
      const position = {
        x: 100 + number * 140, // 动态随机定位
        y: 100 + number * 60,
      };
      _state.graph.nodes = [
        {
          id: currentStage, // NOTE: stage id
          title: _.get(values, `${currentStage}.name`),
          type: STAGE,
          ...position,
        },
        ...graph.nodes,
      ];
      _state.nodePosition[currentStage] = position;
      setFieldValue('stages', [...stages, currentStage]);
    }

    this.setState(_state);
  };

  getStageId = array => {
    const number = array.map(o => o.split('_')[1]);
    const max = number.sort(function(a, b) {
      return b - a;
    })[0];
    return max * 1 || 0;
  };

  addStartNode = () => {
    const { setFieldValue, values } = this.props;
    // show Drawer
    this.setState({ visible: true, selected: null });
    // TODO(qme): stage id random
    const stageId = `stage_${this.getStageId(_.get(values, 'stages')) + 1}`;
    setFieldValue('currentStage', stageId);
    setFieldValue(stageId, {
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
    });
  };

  /*
   * Handlers/Interaction
   */

  // Called by 'drag' handler, etc..
  // to sync updates from D3 with the graph
  onUpdateNode = viewNode => {
    const graph = this.state.graph;
    const i = this.getNodeIndex(viewNode);
    graph.nodes[i] = viewNode;

    this.setState({ graph });
  };

  // Node 'mouseUp' handler
  onSelectNode = viewNode => {
    const { nodePosition } = this.state;
    // Deselect events will send Null viewNode
    let state = { selected: viewNode, nodePosition };
    const nodeId = _.get(viewNode, 'id');
    const moved =
      _.get(nodePosition, `${nodeId}.x`) !== _.get(viewNode, 'x') ||
      _.get(nodePosition, `${nodeId}.y`) !== _.get(viewNode, 'y');
    if (viewNode && !moved) {
      state.visible = true;
    } else {
      state.nodePosition[nodeId] = _.pick(viewNode, ['x', 'y']);
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
    const graph = this.state.graph;
    // Delete any connected edges
    const newEdges = graph.edges.filter((edge, i) => {
      return (
        edge.source !== viewNode[NODE_KEY] && edge.target !== viewNode[NODE_KEY]
      );
    });
    graph.nodes = nodeArr;
    graph.edges = newEdges;
    this.setState({ graph, selected: null });
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
    const { setStageDepned } = this.props;
    // console.log('sourceViewNode', sourceViewNode, targetViewNode);
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
      setStageDepned(graph.edges);
    }
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
    const graph = this.state.graph;
    graph.edges = edges;
    this.setState({
      graph,
      selected: null,
    });
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
    const { nodes, edges } = this.state.graph;
    const selected = this.state.selected;
    const { NodeTypes, NodeSubtypes, EdgeTypes } = GraphConfig;
    const { values } = this.props;

    return (
      <div id="graph" className={styles['graph']}>
        <div className="graph-header">
          <Button type="primary" onClick={this.addStartNode}>
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
        />
        <Drawer
          title="Basic Drawer"
          placement="right"
          closable={false}
          onClose={this.onClose}
          visible={this.state.visible}
          width={600}
        >
          <AddStage
            key={_.get(values, 'currentStage')}
            setFieldValue={this.props.setFieldValue}
            values={this.props.values}
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
};

export default Graph;
