import * as React from 'react';
import { Spin } from 'antd';
import { GraphView } from 'react-digraph';
import GraphConfig, {
  NODE_KEY,
  STAGE,
  SPECIAL_EDGE_TYPE,
} from '../add/graph-config'; // Configures node/edge types
import classNames from 'classnames';
import PropTypes from 'prop-types';

class StageDepend extends React.Component {
  GraphView;

  constructor(props) {
    super(props);

    const { stages } = props;
    let graph = {
      edges: [],
      nodes: [],
    };
    if (!_.isEmpty(stages)) {
      graph = this.tranformStage(stages);
    }
    this.state = {
      graph,
      selected: null,
    };

    this.GraphView = React.createRef();
  }

  componentDidUpdate(preProps) {
    const { stages } = this.props;
    if (!_.isEmpty(stages) && !_.isEqual(stages, preProps.stages)) {
      const data = this.tranformStage(stages);
      this.setState({ graph: data });
    }
  }

  tranformStage = stages => {
    let nodes = [];
    let edges = [];
    _.forEach(stages, (v, k) => {
      const node = {
        id: `stage_${k + 1}`,
        title: v.name,
        type: STAGE,
        x: _.get(nodes, `[${k - 1}].x`, 0) + ((k + 1) % 2) * 240,
        y: _.get(nodes, `[${k - 1}].y`, 0) + (k % 2) * 120,
      };
      nodes.push(node);
      if (_.isArray(v.depends)) {
        const edge = _.map(v.depends, d => {
          return {
            source: d,
            target: `stage_${k + 1}`,
            type: SPECIAL_EDGE_TYPE,
          };
        });
        edges = _.concat(edges, edge);
      }
    });
    return { nodes, edges };
  };

  // Node 'mouseUp' handler
  onSelectNode = viewNode => {
    // Deselect events will send Null viewNode
    let state = { selected: viewNode };
    this.setState(state);
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
    if (nodes.length === 0) {
      return <Spin />;
    }

    return (
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
        renderNodeText={this.renderNodeText}
        readOnly
        showGraphControls={false}
        maxZoom={1}
      />
    );
  }
}

StageDepend.propTypes = {
  values: PropTypes.object,
  setStageDepned: PropTypes.func,
  stages: PropTypes.object,
};

export default StageDepend;
