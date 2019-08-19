import tensorflow as tf
from tensorflow.python.saved_model import signature_constants
from tensorflow.python.saved_model import tag_constants

export_dir = './00000002'
graph_pb = './creditcardfraud.pb'

builder = tf.saved_model.builder.SavedModelBuilder(export_dir)

with tf.gfile.GFile(graph_pb, "rb") as f:
    graph_def = tf.GraphDef()
    graph_def.ParseFromString(f.read())

sigs = {}

with tf.Session(graph=tf.Graph()) as sess:
    # name="" is important to ensure we don't get spurious prefixing
    tf.import_graph_def(graph_def, name="")
    g = tf.get_default_graph()
    inp1 = g.get_tensor_by_name("transaction:0")
    inp2 = g.get_tensor_by_name("reference:0")
    out = g.get_tensor_by_name("output:0")

    sigs[signature_constants.DEFAULT_SERVING_SIGNATURE_DEF_KEY] = \
        tf.saved_model.signature_def_utils.predict_signature_def(
            {"transaction": inp1,"reference" :inp2 }, {"output": out})

    builder.add_meta_graph_and_variables(sess,
                                         [tag_constants.SERVING] ,
                                         signature_def_map=sigs)

builder.save()