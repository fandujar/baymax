import json
import logging
import os
import subprocess
from flask import Flask
from flask import request
from flask import Response
from werkzeug import ImmutableMultiDict

app = Flask(__name__)

def exec_command(plugin: dict, arguments:ImmutableMultiDict=None):
    command_list = []
    command_list.append(plugin['command'])

    for required_arguments in plugin['arguments']['required']:
        '''to-do implement logic to check if required arguments are been informed'''
        pass

    if not arguments:
        command_list.append(arguments)
    process = subprocess.run([plugin['command'],arguments], 
                         stdout=subprocess.PIPE, 
                         universal_newlines=True)
    return process.stdout

@app.route('/')
def plugin():
    plugin_list = PluginList()
    return Response(response=str(plugin_list.value), status=200, mimetype='application/json')

@app.route('/plugin/execute/<plugin_name>', methods = ['GET'])
def execute_plugin(plugin_name:str):
    
    plugin_list = PluginList()
    for plugin in plugin_list.value['plugins']:
        if plugin['name'] == plugin_name:
            arguments = request.args
            print(arguments)
            output = "output"
            # output = exec_command(plugin=plugin, arguments=arguments)
            return Response(response=output, status=200, mimetype='application/json')

    return Response(status=404, response='plugin not found', mimetype='application/json')

@app.route('/plugin/refresh', methods = ['POST'])
def refresh_plugins():
    
    data = []
    for plugin_folder in os.listdir('scripts'):
        with open(f'scripts/{plugin_folder}/plugin.json') as json_file:
            data.append(json.load(json_file))

    plugin_list = PluginList(value=data)
    print(plugin_list.value)
    return Response(status=200, mimetype='application/json')

class PluginList(object):
    __instance = None
    def __new__(cls, value=None):
        if PluginList.__instance is None:
            PluginList.__instance = object.__new__(cls)
            PluginList.__instance.value = {'plugins': []}
        
        if not value == None:
            PluginList.__instance.value['plugins'] = value

        return PluginList.__instance

app.run()


