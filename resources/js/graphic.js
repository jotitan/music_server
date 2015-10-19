/* Create and display specific graphic */

var GRAPH_REFRESH_TIME = 3000;

if(Loader){Loader.toLoad("html/graph.html");}

var GraphicAction = {
    panel:null,
    graph:null,
    init:function(){
        this.panel = $('#addGraphPanel')
        $.extend(this,Panel)
        this.initPanel(this.panel,'Create graph')

        $('.create_graph',this.panel).unbind('click').bind('click',function(){
            GraphicAction.create();
        })
        this.panel.find('#chk_node_all').bind('click',function(){
            GraphicAction.panel.find('.nodes_list :checkbox').prop('checked',$(this).is(':checked'))
        })
    } ,
    showPanel:function(){
        this.panel.find('.metric_select,.nodes_list').empty()
        var data;
        for(var node in ClusterAction.currentDatas){
            data = ClusterAction.currentDatas[node];
            var id = getId(node)
            var urlNode = ClusterAction.currentDatas[node].ID
            if(urlNode != "" ){
                var name = ClusterAction.currentDatas[node].Alias || urlNode
                this.panel.find('.nodes_list').append('<label style="display:block" for="chk_node_' + id + '">'
                    + '<input data-node="' + name + '" data-url="' + urlNode + '" type="checkbox" id="chk_node_'
                    + id + '"/>' + name + '</label>')
            }
        }
        Object.keys(data).filter(function(n){return typeof data[n] == 'number';}).sort().forEach(function(field){
            GraphicAction.panel.find('.metric_select').append('<option>' + field + '</option>')
        })
        $('#chk_node_all').removeAttr('checked','')
        this.open()
    },
    _defineUnit:function(field){
        if(field.toLowerCase().indexOf("memory")!=-1){
            return "Mb"
        }
        if(field.toLowerCase().indexOf("cpu")!=-1) {
            return "%"
        }
        return ""
    },
    create:function(){
        // Get param and nodes
        var field = this.panel.find('.metric_select').val()
        var nodes = GraphicAction.panel.find('.nodes_list :checkbox:checked')
                    .map(function(pos,element){return {name:$(element).data('node'),url:$(element).data('url')}})
        var type = $(':radio[name="type_graph"]:checked',this.panel).val()
        var datasManager = []
        nodes.each(function(i,node){
            var isMemory = field.toLowerCase().indexOf("memory")!=-1
            var getter = function(){
                value = ClusterAction.currentDatas[getId(node.url)][field]

                return isMemory ? value/1024:value
            }
            datasManager.push({name:node.name,get:getter})
        })

        this.graph = CreateGraphicFromScratch(GRAPH_REFRESH_TIME,datasManager,this._formatTitle(field),this._defineUnit(field),type)
        this.graph.run()
        this.close()
        WindowsNavManager.add(this.graph)
    },
    _formatTitle:function(title){
        return title.replace(/[A-Z]/g,function(){
            return " " + arguments[0].toLowerCase()
        })
    },
    close:function(){
        if(this.graph!=null){
        this.graph.stop()
        this.graph = null;
      }
    }
}

/* Create a graphic and create a new div to display graphic inside */
/* @param datasManager : list with name and get() */
function CreateGraphicFromScratch(frequency,datasManager,title,unit,type){
    var div = CloneDiv('idGraphTemplate','idGraph_')
    $('.title > span:first',div).html("Graphic : " + title)

    return new DynamicGraphic(datasManager,frequency,title.trim(),unit,div,type)
}

// Implement Panel
function DynamicGraphic(datasManager,frequency,title,unit,div,type){
    this.frequency = frequency;
    this.datasManager = datasManager;
    this.inverval = null;
    this.graph = $('div.graph',div);
    this.type = type;   // line or pie

    this.run = function(){
        var _self = this;
        this.interval = setInterval(function(){
            _self.updateData();
        },this.frequency)
        this.div.bind('close',function(){
            _self.stop();
        })
    }

    this.updateData = function(){
        this.types.current.update(this.graph);
    }

    this.types = {
        current:null,
        line:{
            update:function(g){
                var shift = g.highcharts().series[0].data.length > 100;
                var time = new Date().getTime()
                this.managers.forEach(function(manager,i){
                    g.highcharts().series[i].addPoint([time,manager.get()],true,shift);
                });
            },
            create:function(managers){
                var series = [];
                managers.forEach(function(manager){
                    series.push({marker:{enabled:false},name:manager.name,data:[]})
                })
                this.managers = managers;
                return {
                    chart:{type:'spline'},
                    tooltip:{
                        shared:true,
                        valueDecimals:2,
                        valueSuffix:" " + unit
                    },
                    title:{text:null},
                    credits:{enabled:false},
                    xAxis:{
                        type:'datetime',
                        maxZoom: 5 * 60 * 1000
                    },
                    yAxis:{
                        'title':{text:title + " (" + unit + ")"},
                        min:0
                    },
                    series:series
                }
            }
        },
        pie:{
           managers:null,
           update:function(g){
                g.highcharts().series[0].data.forEach(function(point,i){
                    point.update(this.managers[point.name]())
                },this);
           },
           create:function(managers){
                var data = [];
                var map = [];
                managers.forEach(function(manager){
                    data.push({name:manager.name,y:manager.get()})
                    map[manager.name] = manager.get
                })
                this.managers = map;
                return {
                    credits:{enabled:false},
                    chart:{
                        plotBackgroundColor: null,
                        plotBorderWidth: null,
                        plotShadow: false,
                        type:'pie'},
                    title:{text:null},
                    plotOptions: {
                        pie: {
                            allowPointSelect: true,
                            cursor: 'pointer',
                            dataLabels: {
                                enabled: false
                            },
                            showInLegend: true
                        }
                    },
                    series: [{
                        name:"Value",
                        colorByPoint: true,
                        data: data
                    }]
                }
            }
        }
    }

    this.stop = function(){
        if(this.interval!=null){
            clearInterval(this.interval)
            this.interval = null;
        }
    }

    this.init = function(){
        Highcharts.setOptions({global:{timezoneOffset:-120}})
        $.extend(this,Panel) ;
        var formatTitle = title.replace(/^[a-z]/,function(){return arguments[0].toUpperCase()})
        this.initPanel(div,'<span class="glyphicon glyphicon-signal"></span> ' + formatTitle);
        this.div.show();

        this.types.current = this.types[this.type];
        var options = this.types.current.create(this.datasManager);
        this.graph.highcharts(options);
    }
    this.init();
}

/* Display cluster memory status in footer */
var BarManager = {
    data:[],
    bar:null,
    titleBloc:null,
    nbByGraph:50,
    formater:null,
    init:function(id,formater){
         this.bar = $('#' + id).peity("line",{width:100,height:20})
         this.titleBloc = $('#' + id).next()
         this.formater = formater;
    },
    update:function(value){
        if(this.bar == null){return;}
        if(this.data.length > this.nbByGraph){
            this.data.shift();
        }
        this.data.push(value)
        this.bar.text(this.data.join(",")).change()
        if(this.formater != null){
            value = this.formater(value)
        }
        this.titleBloc.attr('title',value)
    }
}