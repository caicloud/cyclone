import * as React from 'react';
import { Spin } from 'antd';
import { GraphView } from 'react-digraph';
import GraphConfig, { NODE_KEY } from '../add/graph-config'; // Configures node/edge types
import { tranformStage } from '@/lib/util';
import classNames from 'classnames';
import PropTypes from 'prop-types';

class StageDepend extends React.Component {
  GraphView;

  constructor(props) {
    super(props);

    const { detail } = props;
    let graph = {
      edges: [],
      nodes: [],
    };
    if (!_.isEmpty(detail)) {
      graph = tranformStage(
        _.get(detail, 'spec.stages'),
        _.get(detail, 'metadata.annotations.stagePosition')
      );
    }
    this.state = {
      graph,
      selected: null,
    };

    this.GraphView = React.createRef();
  }

  componentDidUpdate(preProps) {
    const { detail } = this.props;
    if (!_.isEmpty(detail) && !_.isEqual(detail, preProps.detail)) {
      const data = tranformStage(
        _.get(detail, 'spec.stages'),
        _.get(detail, 'metadata.annotations.stagePosition')
      );
      this.setState({ graph: data });
    }
  }

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
      <div style={{ height: '300px' }}>
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
      </div>
    );
  }
}

StageDepend.propTypes = {
  values: PropTypes.object,
  setStageDepned: PropTypes.func,
  stages: PropTypes.object,
  detail: PropTypes.object,
};

export default StageDepend;
