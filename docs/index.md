---
title: Bakery
nav_order: 1
has_children: true
has_toc: false
---

# Bakery

Manifest Manipulation Service


## What's new

Follow the updates to our API here!

<ul>
  {% for post in site.posts limit:15 %}
	{% if post.category contains "whats-new" %}
    <li>
      <span class="post-date">{{ post.date | date: "%B %d, %Y" }}</span> <a href="{{ site.baseurl }}{{ post.url }}">{{ post.title }}</a>
    </li>
    {% endif %}
  {% endfor %}
</ul>

## Articles

<ul>
  {% for post in site.posts limit:15 %}
	{% if post.category contains "article" %}
    <li>
      <span class="post-date">{{ post.date | date: "%B %d, %Y" }}</span> <a href="{{ site.baseurl }}{{ post.url }}">{{ post.title }}</a>
    </li>
    {% endif %}
  {% endfor %}
</ul>


## Help

You can find the source code for Bakery at GitHub:
[bakery][bakery]

[bakery]: https://github.com/cbsinteractive/bakery

If you have any questions regarding Bakery, please reach out in the [#i-vidtech-mediahub](slack://channel?team={cbs}&id={i-vidtech-mediahub}) channel.