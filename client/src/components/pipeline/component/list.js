import React from "react";
import { Table } from "antd";
import { inject, observer } from "mobx-react";
@inject("pipeline")
@observer
class List extends React.Component {
  componentDidMount() {
    this.props.pipeline.getPipelineList();
  }
  render() {
    const { pipeline } = this.props;
    console.log("pipeline", pipeline.pipelineList);
    const columns = [
      {
        title: "Name",
        dataIndex: "name",
        key: "name"
      },
      { title: "Alias", dataIndex: "alias", key: "alias" }
    ];
    return <Table columns={columns} dataSource={pipeline.pipelineList} />;
  }
}

export default List;
