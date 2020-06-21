import React from 'react';
import QueueAnim from 'rc-queue-anim';
import TweenOne from 'rc-tween-one';
import { Row, Col,Card,Form,Space, Input,Button, Modal, Progress } from 'antd';
import OverPack from 'rc-scroll-anim/lib/ScrollOverPack';

class Content1 extends React.Component {

  state={
    url:"",
    share:0,
    cmt:0,
    play:0,
    username:"",
    cookies:"",
    modelVisible:false,
    finalLink:"",
    progress:0
  }
   layout = {
    labelCol: { span: 6 },
    wrapperCol: { span: 24 },
  };
  changeUrl = (event)=>{
    this.setState({
      url:event.target.value
    })
  }
  changeShare = (event)=>{
    this.setState({
      share: Number(event.target.value)
    })
  }
  changeCmt = (event)=>{
    this.setState({
      cmt: Number(event.target.value)
    })
  }
  changePlay = (event)=>{
    this.setState({
      play: Number(event.target.value)
    })
  }
  changeCookies=(event)=>{
    this.setState({
      cookies: event.target.value
    })
  }
  changeUsername=(event)=>{
    this.setState({
      username: event.target.value
    })
  }
  handleTiktok=(kind)=>{
    let cookiesObj = {}
    
    try {
      cookiesObj = JSON.parse(this.state.cookies)
      console.log(cookiesObj)
    } catch (error) {
      
    }
    let data = {
      username:this.state.username,
      cookies:JSON.stringify(cookiesObj),
      url:this.state.url,
      kind:kind,
      share:this.state.share,
      cmt: this.state.cmt,
      play:this.state.play
    }
    console.log(JSON.stringify(data))
    let ws = new WebSocket("ws://localhost:8080/tiktok")
    ws.onopen  = (event)=>{
      ws.send(JSON.stringify(data))
    }
    this.setState({
      modelVisible:true
    })
    ws.onmessage= (mess)=>{
      console.log(mess)
      let res = JSON.parse(mess.data)
      console.log(res)
      if(res.state===0){
        this.setState({
          modelVisible:false
        })
      }
      else if (res.result == ""){
        this.setState({
          progress: Math.floor((res.progress/res.total)*100)
        })
      } else{
        console.log(res.result)
        window.location.href=res.result
        this.setState({
          progress:100,
          finalLink:res.result,
          modelVisible:false    
        })
        ws.close()
        this.setState({
          modelVisible:false
        })
      }
    }
  }
  renderTiktok = ()=>{
    return(
      <Form {...this.layout}>
        <Form.Item
        label="Link Profile"
        name="url"
        rules={[{ required: true, message: 'Please input profile link!' }]}
      >
        <Input onChange={this.changeUrl} />
      </Form.Item>
      <Form.Item
        label="Filter"
        name="filter"
      >
        <Row>
          <Col span={7} style={{marginLeft:"5px"}}>
          <Input onChange={this.changeCmt} type="number" placeholder="Number comment " />
          </Col>
          <Col span={7} style={{marginLeft:"5px"}}>
          <Input onChange={this.changePlay} type="number" placeholder="Number play count"/>
          </Col>
          <Col span={7} style={{marginLeft:"5px"}}>
          <Input onChange={this.changeShare} type="number" placeholder="Number share"/>
          </Col>
        </Row>
      </Form.Item>
      <Form.Item style={{textAlign:"center"}}>
      <Button type="primary" onClick={()=>this.handleTiktok(0)}>
          Download
        </Button>
      </Form.Item>
      </Form>
    )
  }
  renderInstagram =()=>{
    return(
      <Form {...this.layout}>
        <Form.Item
        label="Link Profile"
        name="url"
        rules={[{ required: true, message: 'Please input profile link!' }]}
      >
        <Input onChange={this.changeUsername} />
      </Form.Item>
      <Form.Item name={['cookies', 'Cookies']} label="Paste Cookies">
        <Input.TextArea onChange={this.changeCookies} />
      </Form.Item>
      <Form.Item style={{textAlign:"center"}}>
      <Button type="primary" onClick={()=>this.handleTiktok(1)}>
          Download
        </Button>
      </Form.Item>
      </Form>
    )
  }
  renderFacebook =()=>{
    return(
      <Form {...this.layout}>
        <Form.Item
        label="Video Link"
        name="url"
        rules={[{ required: true, message: 'Please input video link!' }]}
      >
        <Input onChange={this.changeUrl} />
      </Form.Item>
      <Form.Item style={{textAlign:"center"}}>
      <Button type="primary" onClick={()=>this.handleTiktok(2)}>
          Download
        </Button>
      </Form.Item>
      </Form>
    )
  }
  renderProgess =(progress)=>{
    return(
      <Card title="This automatic download when done. Please wait!">
        <Progress
      strokeColor={{
        from: '#108ee9',
        to: '#87d068',
      }}
      percent={progress}
      status="active"
    />
      </Card>
    )
  }
  render() {
  const { dataSource, isMobile, kind, hide } = this.props;
  const animType = {
    queue: isMobile ? 'bottom' : 'right',
    one: isMobile
      ? {
        scaleY: '+=0.3',
        opacity: 0,
        type: 'from',
        ease: 'easeOutQuad',
      }
      : {
        x: '-=30',
        opacity: 0,
        type: 'from',
        ease: 'easeOutQuad',
      },
  };
  return (
    hide === true ? null:
    <div  {...this.props} {...dataSource.wrapper}>
      <Modal
          title="Almost done ..."
          visible={this.state.modelVisible}
          confirmLoading={true}
          cancelButtonProps={{disabled:true}}
        >
         {this.renderProgess(this.state.progress)}
        </Modal>
      <OverPack {...dataSource.OverPack} component={Row}>
        <TweenOne
          key="img"
          animation={animType.one}
          resetStyle
          {...dataSource.imgWrapper}
          component={Col}
          componentProps={{
            md: dataSource.imgWrapper.md,
            xs: dataSource.imgWrapper.xs,
          }}
        >
          <span {...dataSource.img}>
            <img src={dataSource.img.children} width="100%" alt="img" />
          </span>
        </TweenOne>
        <QueueAnim
          key="text"
          type={animType.queue}
          leaveReverse
          ease={['easeOutQuad', 'easeInQuad']}
          {...dataSource.textWrapper}
          component={Col}
          componentProps={{
            md: dataSource.textWrapper.md,
            xs: dataSource.textWrapper.xs,
          }}
        >
          <h2 key="h1" {...dataSource.title}>
            {dataSource.title.children}
          </h2>
          {kind === 0?this.renderTiktok():null}
          {kind === 1? this.renderInstagram():null}
          {kind === 2? this.renderFacebook():null}
        </QueueAnim>
      </OverPack>
    </div>
  );
}
}
export default Content1;
