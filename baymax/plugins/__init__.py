import json
import logging
import os
import subprocess
from flask import request
from flask import Response
from flask import Blueprint
from werkzeug.datastructures import ImmutableMultiDict

baymax_plugins = Blueprint(
        name="baymax_plugins",
        import_name=__name__
    )

plugins_dir = 'baymax/plugins/scripts'

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

@baymax_plugins.route('/plugin')
def plugin():
    plugin_list = PluginList()
    return Response(response=str(plugin_list.value), status=200, mimetype='application/json')

@baymax_plugins.route('/plugin/execute/<plugin_name>', methods = ['GET'])
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

@baymax_plugins.route('/plugin/refresh', methods = ['POST'])
def refresh_plugins():
    data = []
    for plugin_folder in os.listdir(plugins_dir):
        if os.path.isdir(f'{plugins_dir}/{plugin_folder}'):
            with open(f'{plugins_dir}/{plugin_folder}/plugin.json') as json_file:
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


